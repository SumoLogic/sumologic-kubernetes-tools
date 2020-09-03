use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use std::{convert::Infallible, net::SocketAddr};

use clap::{value_t, App, Arg};
use hyper::server::conn::AddrStream;
use hyper::service::{make_service_fn, service_fn};
use hyper::Server;

mod metrics;
mod router;
mod statistics;
use statistics::Statistics;
mod time;

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
      .arg(Arg::with_name("print_headers")
          .long("print-headers")
          .value_name("print_headers")
          .help("Use to print received requests' headers")
          .takes_value(false)
          .required(false))
      .arg(Arg::with_name("print_metrics")
          .short("m")
          .long("print-metrics")
          .value_name("print_metrics")
          .help("Use to print received metrics (with dimensions) on stdout")
          .takes_value(false)
          .required(false))
      .get_matches();

    let port = value_t!(matches, "port", u16).unwrap_or(3000);
    let hostname = value_t!(matches, "hostname", String).unwrap_or("localhost".to_string());
    let print_logs = matches.is_present("print_logs");
    let print_headers = matches.is_present("print_headers");
    let print_metrics = matches.is_present("print_metrics");

    let stats = Statistics {
        metrics: 0,
        logs: 0,
        logs_bytes: 0,
        p_metrics: 0,
        p_logs: 0,
        p_logs_bytes: 0,
        ts: time::get_now(),
        metrics_list: HashMap::new(),
        metrics_ip_list: HashMap::new(),
        logs_ip_list: HashMap::new(),
        url: format!("http://{}:{}/receiver", hostname, port),
        print_logs: print_logs,
        print_headers: print_headers,
        print_metrics: print_metrics,
    };
    let statistics = Arc::new(Mutex::new(stats));

    run_app(statistics, port).await;
}

async fn run_app(stats: Arc<Mutex<Statistics>>, port: u16) {
    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    println!("Receiver mock is waiting for enemy on 0.0.0.0:{}!", port);
    let make_svc = make_service_fn(|conn: &AddrStream| {
        let statistics = stats.clone();
        let address = conn.remote_addr().ip();
        async move {
            let statistics = statistics.clone();
            let result = service_fn(move |req| router::handle(req, address, statistics.clone()));
            Ok::<_, Infallible>(result)
        }
    });

    let server = Server::bind(&addr).serve(make_svc);

    if let Err(e) = server.await {
        eprintln!("server error: {}", e);
    }
}
