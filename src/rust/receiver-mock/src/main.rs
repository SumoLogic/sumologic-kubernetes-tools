use std::{convert::Infallible, net::SocketAddr};
use hyper::{Body, Request, Response, Server};
use hyper::service::{make_service_fn, service_fn};
use hyper::header::HeaderValue;
use std::time::{SystemTime, UNIX_EPOCH};
use std::sync::Arc;
use std::sync::Mutex;
use std::vec::Vec;
use std::collections::HashMap;
use clap::{Arg, App, value_t};

fn get_now() -> u64 {
  let start = SystemTime::now();
  let since_the_epoch = start.duration_since(UNIX_EPOCH)
      .expect("Time went backwards");
  return since_the_epoch.as_secs() as u64;
}

struct Statistics {
  metrics: u64,
  logs: u64,
  logs_bytes: u64,
  p_metrics: u64,
  p_logs: u64,
  p_logs_bytes: u64,
  ts: u64,
  metrics_list: HashMap<String, u64>,
}

async fn handle(req: Request<Body>, statistics: Arc<Mutex<Statistics>>, port: u16) -> Result<Response<Body>, Infallible> {
    let uri = req.uri().path();
    
    match uri {
      "/metrics-list" => {
        let statistics = statistics.lock().unwrap();
        let mut string = "".to_string();
        for metric in (*statistics).metrics_list.iter() {
          string.push_str(&format!("{}: {}\n", &metric.0, &metric.1));
        }
        // ToDo: Do it properly, like human in json
        Ok(Response::new(format!("{}", string).into()))
      }
      "/metrics" => {
        let statistics = statistics.lock().unwrap();
  
        Ok(Response::new(format!("# TYPE receiver_mock_metrics_count counter
receiver_mock_metrics_count {}
# TYPE receiver_mock_logs_count counter
receiver_mock_logs_count {}
# TYPE receiver_mock_logs_bytes_count counter
receiver_mock_logs_bytes_count {}",
          (*statistics).metrics,
          (*statistics).logs,
          (*statistics).logs_bytes).into()))
      },
      "/metrics-reset" => {
        let mut statistics = statistics.lock().unwrap();
        for (_, val) in (*statistics).metrics_list.iter_mut() {
          *val = 0;
        }
        Ok(Response::new("All counters reset successfully".into()))
      }
      _ => {
        if uri.starts_with("/terraform") {
          // ToDo: Do it properly, like human
          Ok(Response::new(format!("{{\"source\": {{\"url\": \"http://receiver-mock.receiver-mock:{}/receiver\"}}}}", port).into()))
        }
        else {
          let empty_header = HeaderValue::from_str("").unwrap();
          let content_type = req.headers().get("content-type").unwrap_or(&empty_header).to_str().unwrap();
          match content_type {
            "application/vnd.sumologic.prometheus" => {
              let whole_body = hyper::body::to_bytes(req.into_body()).await.unwrap();
              let mut statistics = statistics.lock().unwrap();
              let vector_body = whole_body.into_iter().collect::<Vec<u8>>();
              let string_body = String::from_utf8(vector_body).unwrap();
              
              let lines = string_body.trim().split("\n");
    
              for line in lines {
                let metric_name = line.split("{").nth(0).unwrap().to_string();
                let saved_metric = (*statistics).metrics_list.entry(metric_name).or_insert(0);
                *saved_metric += 1;
                (*statistics).metrics += 1;
              }
            },
            "application/x-www-form-urlencoded" => {
              let whole_body = hyper::body::to_bytes(req.into_body()).await.unwrap();
              let vector_body = whole_body.into_iter().collect::<Vec<u8>>();
    
              let mut statistics = statistics.lock().unwrap();
              (*statistics).logs_bytes += vector_body.len() as u64;
    
              let string_body = String::from_utf8(vector_body).unwrap();
              (*statistics).logs += string_body.trim().split("\n").count() as u64;
            },
            &_ => {
              println!("invalid header value");
            }
          }
          stats(statistics);
          Ok(Response::new("".into()))
        }
      }
    }
}

async fn run_app(statistics: Arc<Mutex<Statistics>>, port: u16) {
    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    println!("Receiver mock is waiting for enemy on 0.0.0.0:{}!", port);
    let make_svc = make_service_fn(|_conn| {
      let statistics = statistics.clone();
      async move {
        let statistics = statistics.clone();
        let result = service_fn(move |req| handle(
          req,
          statistics.clone(),
          port,
        ));
        Ok::<_, Infallible>(result)
    }});

    let server = Server::bind(&addr).serve(make_svc);

    if let Err(e) = server.await {
        eprintln!("server error: {}", e);
    }

}

fn stats(statistics: Arc<Mutex<Statistics>>) {
  let mut statistics = statistics.lock().unwrap();

  if get_now() >= (*statistics).ts + 60 {
      println!("{} Metrics: {:10.} Logs: {:10.}; {:6.6} MB/s",
        (*statistics).ts,
        (*statistics).metrics - (*statistics).p_metrics,
        (*statistics).logs - (*statistics).p_logs,
        (((*statistics).logs_bytes - (*statistics).p_logs_bytes) as f64)/((get_now()-(*statistics).ts) as f64)/(1e6 as f64));
      (*statistics).ts = get_now();
      (*statistics).p_metrics = (*statistics).metrics;
      (*statistics).p_logs = (*statistics).logs;
      (*statistics).p_logs_bytes = (*statistics).logs_bytes;
  }
}

#[tokio::main]
pub async fn main() {
    let matches = App::new("Receiver mock")
      .version("0.0")
      .author("Dominik Rosiek <drosiek@sumologic.com>")
      .about("Receiver mock can be used for testing performance or functionality of kubernetes collection without sending data to sumologic")
      .arg(Arg::with_name("port")
          .short("p")
          .long("port")
          .value_name("port")
          .help("Port to listen")
          .takes_value(true)
          .required(false))
      .get_matches();

    let port: u16 = value_t!(matches, "port", u16).unwrap_or(3000);

    let statistics = Statistics {
      metrics: 0,
      logs: 0,
      logs_bytes: 0,
      p_metrics: 0,
      p_logs: 0,
      p_logs_bytes: 0,
      ts: get_now(),
      metrics_list: HashMap::new(),
    };
    let statistics = Arc::new(Mutex::new(statistics));

    run_app(statistics, port).await;
}