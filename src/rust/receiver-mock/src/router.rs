use std::collections::HashMap;
use std::net::{IpAddr, Ipv4Addr};
use std::sync::Mutex;

use actix_http::http;
use actix_web::{web, HttpRequest, HttpResponse, Responder};
use bytes;
use chrono::Duration;
use serde_derive::Serialize;

use crate::metrics;
use crate::options;
use crate::time::get_now;

pub struct AppState {
    // Mutexes are necessary to mutate data safely across threads in handlers.
    //
    pub metrics: Mutex<u64>,
    pub logs: Mutex<u64>,
    pub logs_bytes: Mutex<u64>,

    pub metrics_list: Mutex<HashMap<String, u64>>,
    pub metrics_ip_list: Mutex<HashMap<IpAddr, u64>>,
    // logs_ip_list: .0 is logs counter, .1 is bytes counter
    pub logs_ip_list: Mutex<HashMap<IpAddr, (u64, u64)>>,
}

impl AppState {
    pub fn add_metrics_result(&self, result: metrics::MetricsHandleResult) {
        {
            let mut metrics = self.metrics.lock().unwrap();
            *metrics += result.metrics;
        }

        {
            let mut metrics_list = self.metrics_list.lock().unwrap();
            for (name, count) in result.metrics_list.iter() {
                *metrics_list.entry(name.clone()).or_insert(0) += count;
            }
        }

        {
            let mut metrics_ip_list = self.metrics_ip_list.lock().unwrap();
            for (&ip_address, count) in result.metrics_ip_list.iter() {
                *metrics_ip_list.entry(ip_address).or_insert(0) += count;
            }
        }
    }
}

pub struct AppMetadata {
    pub url: String,
}

// Reset metrics counter
pub async fn handler_metrics_reset(app_state: web::Data<AppState>) -> impl Responder {
    app_state.metrics_list.lock().unwrap().clear();
    app_state.metrics_ip_list.lock().unwrap().clear();

    HttpResponse::Ok().body("All counters reset successfully")
}

// List metrics in format: <name>: <count>
pub async fn handler_metrics_list(app_state: web::Data<AppState>) -> impl Responder {
    let mut out = String::new();
    let metrics_list = app_state.metrics_list.lock().unwrap();
    for (name, count) in metrics_list.iter() {
        out.push_str(&format!("{}: {}\n", name, count));
    }
    HttpResponse::Ok().body(out)
}

// List metrics in format: <ip_address>: <count>
pub async fn handler_metrics_ips(app_state: web::Data<AppState>) -> impl Responder {
    let mut out = String::new();
    let metrics_ip_list = app_state.metrics_ip_list.lock().unwrap();
    for (ip_address, count) in metrics_ip_list.iter() {
        out.push_str(&format!("{}: {}\n", ip_address, count));
    }
    HttpResponse::Ok().body(out)
}

// Metrics in prometheus format
pub async fn handler_metrics(app_state: web::Data<AppState>) -> impl Responder {
    let mut body = format!(
        "# TYPE receiver_mock_metrics_count counter
receiver_mock_metrics_count {}
# TYPE receiver_mock_logs_count counter
receiver_mock_logs_count {}
# TYPE receiver_mock_logs_bytes_count counter
receiver_mock_logs_bytes_count {}\n",
        app_state.metrics.lock().unwrap(),
        app_state.logs.lock().unwrap(),
        app_state.logs_bytes.lock().unwrap(),
    );

    {
        let metrics_ip_list = app_state.metrics_ip_list.lock().unwrap();
        if metrics_ip_list.len() > 0 {
            let mut metrics_ip_string =
                String::from("# TYPE receiver_mock_metrics_ip_count counter\n");
            for (ip, count) in metrics_ip_list.iter() {
                metrics_ip_string.push_str(&format!(
                    "receiver_mock_metrics_ip_count{{ip_address=\"{}\"}} {}\n",
                    ip, count
                ));
            }
            body.push_str(&metrics_ip_string);
        }
    }

    {
        let logs_ip_list = app_state.logs_ip_list.lock().unwrap();
        if logs_ip_list.len() > 0 {
            let mut logs_ip_count_bytes_string =
                String::from("# TYPE receiver_mock_logs_bytes_ip_count counter\n");
            let mut logs_ip_count_string =
                String::from("# TYPE receiver_mock_logs_ip_count counter\n");

            for (ip, val) in logs_ip_list.iter() {
                logs_ip_count_string.push_str(&format!(
                    "receiver_mock_logs_ip_count{{ip_address=\"{}\"}} {}\n",
                    ip, val.0
                ));
                logs_ip_count_bytes_string.push_str(&format!(
                    "receiver_mock_logs_bytes_ip_count{{ip_address=\"{}\"}} {}\n",
                    ip, val.1
                ));
            }
            body.push_str(&logs_ip_count_string);
            body.push_str(&logs_ip_count_bytes_string);
        }
    }

    HttpResponse::Ok().body(body)
}

pub async fn handler_terraform(app_metadata: web::Data<AppMetadata>) -> impl Responder {
    #[derive(Serialize)]
    struct TerraformResponse {
        source: String,
    }

    web::Json(TerraformResponse {
        source: app_metadata.url.clone(),
    })
}

pub async fn handler_receiver(
    req: HttpRequest,
    body: bytes::Bytes,
    app_state: web::Data<AppState>,
    opts: web::Data<options::Options>,
) -> impl Responder {
    // Don't fail when we can't read remote address.
    // Default to localhost and just ingest what was sent.
    let localhost: std::net::SocketAddr =
        std::net::SocketAddr::new(IpAddr::V4(Ipv4Addr::new(127, 0, 0, 1)), 0);
    let remote_address = req.peer_addr().unwrap_or(localhost).ip();

    let body_length = body.len() as u64;
    // actix automatically decompresses body for us.
    let string_body = String::from_utf8(body.to_vec()).unwrap();
    let lines = string_body.trim().lines();

    let headers = req.headers();
    if opts.print.headers {
        print_request_headers(req.method(), req.version(), req.uri(), headers);
    }
    let empty_header = http::HeaderValue::from_str("").unwrap();
    let content_type = headers
        .get("content-type")
        .unwrap_or(&empty_header)
        .to_str()
        .unwrap();

    match content_type {
        // Metrics in carbon2 format
        "application/vnd.sumologic.carbon2" => {
            let result = metrics::handle_carbon2(lines, remote_address, opts.print);
            app_state.add_metrics_result(result);
        }

        // Metrics in graphite format
        "application/vnd.sumologic.graphite" => {
            let result = metrics::handle_graphite(lines, remote_address, opts.print);
            app_state.add_metrics_result(result);
        }

        // Metrics in prometheus format
        "application/vnd.sumologic.prometheus" => {
            let result = metrics::handle_prometheus(lines, remote_address, opts.print);
            app_state.add_metrics_result(result);
        }

        // Logs & events
        "application/x-www-form-urlencoded" => {
            // TODO: refactor
            let mut lines_count = 0 as u64;
            if opts.print.logs {
                for line in lines {
                    println!("log => {}", line);
                    lines_count += 1;
                }
            } else {
                lines_count = lines.count() as u64;
            }

            {
                let mut logs_bytes = app_state.logs_bytes.lock().unwrap();
                *logs_bytes += body_length;
            }

            {
                let mut logs = app_state.logs.lock().unwrap();
                *logs += lines_count;
            }

            {
                let mut logs_ip_list = app_state.logs_ip_list.lock().unwrap();
                let (logs_count, bytes_count) =
                    logs_ip_list.entry(remote_address).or_insert((0, 0));
                *logs_count += lines_count;
                *bytes_count += body_length;
            }
        }

        &_ => {
            println!("invalid header value");
        }
    }

    HttpResponse::Ok()
}

fn print_request_headers(
    method: &http::Method,
    version: http::Version,
    uri: &http::Uri,
    headers: &http::HeaderMap,
) {
    let method = method.as_str();
    let uri = uri.path();

    println!("--> {} {} {:?}", method, uri, version);
    for (key, value) in headers {
        println!("--> {}: {}", key, value.to_str().unwrap());
    }
    println!();
}

// TODO: extract stdout as parameter to make testing easier.
// ref: https://github.com/SumoLogic/sumologic-kubernetes-tools/issues/58
pub fn start_print_stats_timer(
    t: &timer::Timer,
    interval: Duration,
    app_state: web::Data<AppState>,
) -> timer::Guard {
    let mut p_metrics: u64 = 0;
    let mut p_logs: u64 = 0;
    let mut p_logs_bytes: u64 = 0;
    let mut ts = get_now();

    t.schedule_repeating(interval, move || {
        let now = get_now();
        let metrics = app_state.metrics.lock().unwrap();
        let logs = app_state.logs.lock().unwrap();
        let logs_bytes = app_state.logs_bytes.lock().unwrap();

        // TODO: make this print metrics per minute (as DPM) and logs
        // per second, regardless of used interval
        // ref: https://github.com/SumoLogic/sumologic-kubernetes-tools/issues/57
        println!(
            "{} Metrics: {:10.} Logs: {:10.}; {:6.6} MB/s",
            now,
            *metrics - p_metrics,
            *logs - p_logs,
            ((*logs_bytes - p_logs_bytes) as f64) / ((now - ts) as f64) / (1e6 as f64)
        );

        ts = now;
        p_metrics = *metrics;
        p_logs = *logs;
        p_logs_bytes = *logs_bytes;
    })
}
