use std::collections::HashMap;
use std::convert::Infallible;
use std::io::prelude::*;
use std::net::IpAddr;
use std::sync::{Arc, Mutex};
use std::vec::Vec;

use flate2::read::GzDecoder;
use hyper::header::HeaderValue;
use hyper::{Body, Request, Response};
use serde_json::json;

use crate::metrics;
use crate::print;
use crate::statistics;
use crate::statistics::Statistics;

pub async fn handle(
    req: Request<Body>,
    address: IpAddr,
    stats: Arc<Mutex<Statistics>>,
    print_opts: print::Options,
) -> Result<Response<Body>, Infallible> {
    let (parts, body) = req.into_parts();

    if print_opts.print_headers {
        print_request_headers(&parts);
    }

    let uri = parts.uri.path();
    match uri {
        // List metrics in format: <name>: <count>
        "/metrics-list" => {
            let stats = stats.lock().unwrap();
            let mut string = "".to_string();
            for metric in (*stats).metrics_list.iter() {
                string.push_str(&format!("{}: {}\n", &metric.0, &metric.1));
            }
            Ok(Response::new(format!("{}", string).into()))
        }
        // List metrics in format: <ip_address>: <count>
        "/metrics-ips" => {
            let statistics = stats.lock().unwrap();
            let mut string = "".to_string();
            for metric in (*statistics).metrics_ip_list.iter() {
                string.push_str(&format!("{}: {}\n", &metric.0, &metric.1));
            }
            Ok(Response::new(format!("{}", string).into()))
        }
        // Metrics in prometheus format
        "/metrics" => {
            let statistics = stats.lock().unwrap();

            let ip_stats = &statistics.metrics_ip_list;
            let mut metrics_ip_string =
                "# TYPE receiver_mock_metrics_ip_count counter\n".to_string();
            for metric in ip_stats.iter() {
                metrics_ip_string.push_str(&format!(
                    "receiver_mock_metrics_ip_count{{ip_address=\"{}\"}} {}\n",
                    &metric.0, &metric.1
                ));
            }

            let ip_stats = &statistics.logs_ip_list;
            let mut logs_ip_count_string =
                "# TYPE receiver_mock_logs_ip_count counter\n".to_string();
            let mut logs_ip_count_bytes_string =
                "# TYPE receiver_mock_logs_bytes_ip_count counter\n".to_string();
            for metric in ip_stats.iter() {
                logs_ip_count_string.push_str(&format!(
                    "receiver_mock_logs_ip_count{{ip_address=\"{}\"}} {}\n",
                    &metric.0,
                    &(metric.1).0
                ));
                logs_ip_count_bytes_string.push_str(&format!(
                    "receiver_mock_logs_bytes_ip_count{{ip_address=\"{}\"}} {}\n",
                    &metric.0,
                    &(metric.1).1
                ));
            }

            Ok(Response::new(
                format!(
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
                )
                .into(),
            ))
        }
        // Reset metrics counter
        "/metrics-reset" => {
            let mut statistics = stats.lock().unwrap();
            (*statistics).metrics_list = HashMap::new();
            (*statistics).metrics_ip_list = HashMap::new();
            Ok(Response::new("All counters reset successfully".into()))
        }
        _ => {
            // Mock receiver
            if uri.starts_with("/terraform") {
                let statistics = stats.lock().unwrap();
                Ok(Response::new(
                    json!({
                      "source": {
                        "url": *statistics.url,
                      }
                    })
                    .to_string()
                    .into(),
                ))
            }
            // Treat every other url as receiver endpoint
            else {
                let headers = &parts.headers;
                let whole_body = hyper::body::to_bytes(body).await.unwrap();
                let vector_body = whole_body.into_iter().collect::<Vec<u8>>();
                let vector_length = vector_body.len() as u64;
                let mut string_body = String::new();

                let empty_header = HeaderValue::from_str("").unwrap();
                let content_encoding = headers
                    .get("content-encoding")
                    .unwrap_or(&empty_header)
                    .to_str()
                    .unwrap();
                if content_encoding == "gzip" {
                    let mut d = GzDecoder::new(&vector_body[..]);
                    d.read_to_string(&mut string_body).unwrap();
                } else {
                    string_body = String::from_utf8(vector_body).unwrap();
                }

                let lines = string_body.trim().lines();

                let content_type = headers
                    .get("content-type")
                    .unwrap_or(&empty_header)
                    .to_str()
                    .unwrap();
                match content_type {
                    // Metrics in carbon2 format
                    "application/vnd.sumologic.carbon2" => {
                        metrics::handle_carbon2(lines, address, &stats, print_opts);
                    }
                    // Metrics in graphite format
                    "application/vnd.sumologic.graphite" => {
                        metrics::handle_graphite(lines, address, &stats, print_opts);
                    }
                    // Metrics in prometheus format
                    "application/vnd.sumologic.prometheus" => {
                        metrics::handle_prometheus(lines, address, &stats, print_opts);
                    }
                    // Logs & events
                    "application/x-www-form-urlencoded" => {
                        if print_opts.print_logs {
                            let mut counter = 0;
                            for line in lines {
                                println!("log => {}", line);
                                counter += 1;
                            }
                            let mut stats = stats.lock().unwrap();
                            (*stats).logs += counter;
                            (*stats).logs_bytes += vector_length;

                            let logs_ip_list =
                                (*stats).logs_ip_list.entry(address).or_insert((0, 0));
                            (*logs_ip_list).0 += counter;
                            (*logs_ip_list).1 += vector_length;
                        } else {
                            let lines_count = lines.count() as u64;
                            let mut stats = stats.lock().unwrap();
                            (*stats).logs_bytes += vector_length;
                            (*stats).logs += lines_count;

                            let logs_ip_list =
                                (*stats).logs_ip_list.entry(address).or_insert((0, 0));
                            (*logs_ip_list).0 += lines_count;
                            (*logs_ip_list).1 += vector_length;
                        }
                    }
                    &_ => {
                        println!("invalid header value");
                    }
                }
                statistics::print(&stats);
                Ok(Response::new("".into()))
            }
        }
    }
}

pub fn print_request_headers(parts: &http::request::Parts) {
    let method = parts.method.as_str();
    let uri = parts.uri.path();
    let headers = &parts.headers;

    println!("--> {} {} {:?}", method, uri, parts.version);
    for header in headers {
        let key = header.0;
        let value = header.1;
        println!("--> {}: {}", key, value.to_str().unwrap());
    }
    println!();
}
