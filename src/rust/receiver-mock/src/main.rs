use std::{convert::Infallible, net::SocketAddr};
use hyper::{Body, Request, Response, Server};
use hyper::service::{make_service_fn, service_fn};
use hyper::header::HeaderValue;
use hyper::server::conn::AddrStream;
use std::time::{SystemTime, UNIX_EPOCH};
use std::sync::Arc;
use std::sync::Mutex;
use std::vec::Vec;
use std::collections::HashMap;
use clap::{Arg, App, value_t};
use serde_json::json;
use std::net::IpAddr;

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
  metrics_ip_list: HashMap<IpAddr, u64>,
  // logs_ip_list: .0 is logs counter, .1 is bytes counter
  logs_ip_list: HashMap<IpAddr, (u64, u64)>,
  url: String,
  print_logs: bool,
}

async fn handle(req: Request<Body>, address: IpAddr, statistics: Arc<Mutex<Statistics>>) -> Result<Response<Body>, Infallible> {
    let uri = req.uri().path();
    
    match uri {
      // List metrics in format: <name>: <count>
      "/metrics-list" => {
        let statistics = statistics.lock().unwrap();
        let mut string = "".to_string();
        for metric in (*statistics).metrics_list.iter() {
          string.push_str(&format!("{}: {}\n", &metric.0, &metric.1));
        }
        Ok(Response::new(format!("{}", string).into()))
      }
      // List metrics in format: <ip_address>: <count>
      "/metrics-ips" => {
        let statistics = statistics.lock().unwrap();
        let mut string = "".to_string();
        for metric in (*statistics).metrics_ip_list.iter() {
          string.push_str(&format!("{}: {}\n", &metric.0, &metric.1));
        }
        Ok(Response::new(format!("{}", string).into()))
      }
      // Metrics in prometheus format
      "/metrics" => {
        let statistics = statistics.lock().unwrap();

        let ip_stats = &statistics.metrics_ip_list;
        let mut metrics_ip_string = "# TYPE receiver_mock_metrics_ip_count counter\n".to_string();
        for metric in ip_stats.iter() {
          metrics_ip_string.push_str(&format!("receiver_mock_metrics_ip_count{{ip_address=\"{}\"}} {}\n", &metric.0, &metric.1));
        }

        let ip_stats = &statistics.logs_ip_list;
        let mut logs_ip_count_string = "# TYPE receiver_mock_logs_ip_count counter\n".to_string();
        let mut logs_ip_count_bytes_string = "# TYPE receiver_mock_logs_bytes_ip_count counter\n".to_string();
        for metric in ip_stats.iter() {
          logs_ip_count_string.push_str(&format!("receiver_mock_logs_ip_count{{ip_address=\"{}\"}} {}\n", &metric.0, &(metric.1).0));
          logs_ip_count_bytes_string.push_str(&format!("receiver_mock_logs_bytes_ip_count{{ip_address=\"{}\"}} {}\n", &metric.0, &(metric.1).1));
        }
  
        Ok(Response::new(format!(
"# TYPE receiver_mock_metrics_count counter
receiver_mock_metrics_count {}
# TYPE receiver_mock_logs_count counter
receiver_mock_logs_count {}
# TYPE receiver_mock_logs_bytes_count counter
receiver_mock_logs_bytes_count {}
{}
{}
{}",
          (*statistics).metrics,
          (*statistics).logs,
          (*statistics).logs_bytes,
          metrics_ip_string,
          logs_ip_count_string,
          logs_ip_count_bytes_string
        ).into()))
      },
      // Reset metrics counter
      "/metrics-reset" => {
        let mut statistics = statistics.lock().unwrap();
        (*statistics).metrics_list = HashMap::new();
        (*statistics).metrics_ip_list = HashMap::new();
        Ok(Response::new("All counters reset successfully".into()))
      }
      _ => {
        // Mock receiver
        if uri.starts_with("/terraform") {
          let statistics = statistics.lock().unwrap();
          Ok(Response::new(
            json!({
              "source": {
                "url": *statistics.url,
              }
            }).to_string().into()
          ))
        }
        // Treat every other url as receiver endpoint
        else {
          let empty_header = HeaderValue::from_str("").unwrap();
          let content_type = req.headers().get("content-type").unwrap_or(&empty_header).to_str().unwrap();
          match content_type {
            // Metrics
            "application/vnd.sumologic.prometheus" => {
              let whole_body = hyper::body::to_bytes(req.into_body()).await.unwrap();
              let mut stats = statistics.lock().unwrap();
              let vector_body = whole_body.into_iter().collect::<Vec<u8>>();
              let string_body = String::from_utf8(vector_body).unwrap();
              
              let lines = string_body.trim().split("\n");
    
              for line in lines {
                let metric_name = line.split("{").nth(0).unwrap().to_string();
                let saved_metric = (*stats).metrics_list.entry(metric_name).or_insert(0);

                *saved_metric += 1;
                (*stats).metrics += 1;

                let metrics_ip_list = (*stats).metrics_ip_list.entry(address).or_insert(0);
                *metrics_ip_list += 1;
              }
            },
            // Logs & events
            "application/x-www-form-urlencoded" => {
              let whole_body = hyper::body::to_bytes(req.into_body()).await.unwrap();
              let vector_body = whole_body.into_iter().collect::<Vec<u8>>();
              let vector_length = vector_body.len() as u64;
              let mut stats = statistics.lock().unwrap();
    
              let string_body = String::from_utf8(vector_body).unwrap();
              let lines = string_body.trim().split("\n");

              if (*stats).print_logs {
                let mut counter = 0;
                for line in lines {
                  println!("log => {}", line);
                  counter += 1;
                }
                (*stats).logs += counter;
                (*stats).logs_bytes += vector_length;

                let logs_ip_list = (*stats).logs_ip_list.entry(address).or_insert((0, 0));
                (*logs_ip_list).0 += counter;
                (*logs_ip_list).1 += vector_length;
              }
              else {
                let lines_count = lines.count() as u64;
                (*stats).logs_bytes += vector_length;
                (*stats).logs += lines_count;

                let logs_ip_list = (*stats).logs_ip_list.entry(address).or_insert((0, 0));
                (*logs_ip_list).0 += lines_count;
                (*logs_ip_list).1 += vector_length;
              }

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
    let make_svc = make_service_fn(|conn: &AddrStream | {
      let statistics = statistics.clone();
      let address = conn.remote_addr().ip();
      async move {
        let statistics = statistics.clone();
        let result = service_fn(move |req| handle(
          req,
          address,
          statistics.clone()
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
          .help("Port to listen on")
          .takes_value(true)
          .required(false))
      .arg(Arg::with_name("hostname")
          .short("l")
          .long("hostname")
          .value_name("hostname")
          .help("Hostname reported as the receiver. For kubernetes it will be '<service name>.<namespace>'")
          .takes_value(true)
          .required(false))
        .arg(Arg::with_name("print_logs")
            .short("r")
            .long("print-logs")
            .value_name("print_logs")
            .help("Use to print received logs on stdout")
            .takes_value(false)
            .required(false))
      .get_matches();

    let port = value_t!(matches, "port", u16).unwrap_or(3000);
    let hostname = value_t!(matches, "hostname", String).unwrap_or("localhost".to_string());
    let print_logs = matches.is_present("print_logs");

    let statistics = Statistics {
      metrics: 0,
      logs: 0,
      logs_bytes: 0,
      p_metrics: 0,
      p_logs: 0,
      p_logs_bytes: 0,
      ts: get_now(),
      metrics_list: HashMap::new(),
      metrics_ip_list: HashMap::new(),
      logs_ip_list: HashMap::new(),
      url: format!("http://{}:{}/receiver", hostname, port),
      print_logs: print_logs,
    };
    let statistics = Arc::new(Mutex::new(statistics));

    run_app(statistics, port).await;
}