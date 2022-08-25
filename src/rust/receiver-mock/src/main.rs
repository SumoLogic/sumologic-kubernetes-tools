#[macro_use]
extern crate json_str;

use std::collections::HashMap;
use std::sync::Mutex;

use actix_web::web;

use chrono::Duration;
use clap::{value_t, App, Arg};
use log::error;
use log::info;
use std::thread;
use std::time as stime;

mod logs;
mod metrics;
mod traces;

mod options;
use options::Options;
mod metadata;
mod router;
mod time;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    simple_logger::init_with_level(log::Level::Debug).unwrap();

    let matches = App::new("Receiver mock")
      .version("0.0")
      .author("Sumo Logic <collection@sumologic.com>")
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
      .arg(Arg::with_name("store_metrics")
          .long("store-metrics")
          .value_name("store_metrics")
          .help("Use to store metrics which will then be returned via /metrics-samples")
          .takes_value(false)
          .required(false))
      .arg(Arg::with_name("store_logs")
          .long("store-logs")
          .value_name("store_logs")
          .help("Use to store log data which can then be queried via /logs/* endpoints")
          .takes_value(false)
          .required(false))
      .arg(Arg::with_name("drop_rate")
          .short("d")
          .long("drop-rate")
          .value_name("drop_rate")
          .help("Use to specify packet drop rate. This is number from 0 (do not drop) to 100 (drop all).")
          .takes_value(true)
          .required(false))
        .arg(Arg::with_name("delay_time")
            .short("t")
            .long("delay-time")
            .value_name("delay_time")
            .help("Use to specify delay time. It mocks request processing time in milliseconds.")
            .takes_value(true)
            .required(false))
      .get_matches();

    let port = value_t!(matches, "port", u16).unwrap_or(3000);
    let drop_rate = value_t!(matches, "drop_rate", i64).unwrap_or(0);
    let delay_time = stime::Duration::from_millis(value_t!(matches, "delay_time", u64).unwrap_or(0));
    let hostname = value_t!(matches, "hostname", String).unwrap_or("localhost".to_string());
    let opts = Options {
        print: options::Print {
            logs: matches.is_present("print_logs"),
            headers: matches.is_present("print_headers"),
            metrics: matches.is_present("print_metrics"),
        },
        drop_rate: drop_rate,
        delay_time: delay_time,
        store_metrics: matches.is_present("store_metrics"),
        store_logs: matches.is_present("store_logs"),
    };

    run_app(hostname, port, opts).await
}

async fn run_app(hostname: String, port: u16, opts: Options) -> std::io::Result<()> {
    let app_state = web::Data::new(router::AppState::new());

    let t = timer::Timer::new();
    // TODO: configure interval?
    // ref: https://github.com/SumoLogic/sumologic-kubernetes-tools/issues/59
    router::start_print_stats_timer(&t, Duration::seconds(60), app_state.clone()).ignore();

    let app_metadata = web::Data::new(router::AppMetadata {
        url: format!("http://{}:{}/receiver", hostname, port),
    });

    let terraform_state = web::Data::new(router::terraform::TerraformState {
        fields: Mutex::new(HashMap::new()),
    });

    info!("Receiver mock is waiting for enemy on 0.0.0.0:{}!", port);
    let result = actix_web::HttpServer::new(move || {
        actix_web::App::new()
            // Middleware printing headers for all handlers.
            // For a more robust middleware implementation (in its own type)
            // one can take a look at https://actix.rs/docs/middleware/
            .wrap_fn(move |req, srv| {
                if opts.print.headers {
                    let headers = req.headers();

                    router::print_request_headers(req.method(), req.version(), req.uri(), headers);
                }

                thread::sleep(opts.delay_time);

                actix_web::dev::Service::call(&srv, req)
            })
            .app_data(app_state.clone()) // Mutable shared state
            .app_data(web::Data::new(opts.clone()))
            .route(
                "/metrics-reset",
                web::post().to(router::metrics_data::handler_metrics_reset),
            )
            .route(
                "/metrics-list",
                web::get().to(router::metrics_data::handler_metrics_list),
            )
            .route(
                "/metrics-ips",
                web::get().to(router::metrics_data::handler_metrics_ips),
            )
            .route(
                "/metrics-samples",
                web::get().to(router::metrics_data::handler_metrics_samples),
            )
            .route("/metrics", web::get().to(router::handler_metrics))
            .route("/logs/count", web::get().to(router::handler_logs_count))
            .service(
                web::scope("/api/v1")
                    .route(
                        "/collector/register",
                        web::post().to(router::api::v1::handler_collector_register),
                    )
                    .route(
                        "/collector/heartbeat",
                        web::post().to(router::api::v1::handler_collector_heartbeat),
                    ),
            )
            .service(
                web::scope("/terraform")
                    .app_data(app_metadata.clone())
                    .app_data(terraform_state.clone())
                    .route(
                        "/api/v1/fields/quota",
                        web::get().to(router::terraform::handler_terraform_fields_quota),
                    )
                    .route(
                        "/api/v1/fields/{field}",
                        web::get().to(router::terraform::handler_terraform_field),
                    )
                    .route(
                        "/api/v1/fields",
                        web::get().to(router::terraform::handler_terraform_fields),
                    )
                    .route(
                        "/api/v1/fields",
                        web::post().to(router::terraform::handler_terraform_fields_create),
                    )
                    .default_service(web::get().to(router::terraform::handler_terraform)),
            )
            .route("/dump", web::post().to(router::handler_dump))
            // OTLP
            .service(
                web::scope("/receiver/v1")
                    .route(
                        "/logs",
                        web::post().to(router::otlp::handler_receiver_otlp_logs),
                    )
                    .route(
                        "/metrics",
                        web::post().to(router::otlp::handler_receiver_otlp_metrics),
                    )
                    .route(
                        "/traces",
                        web::post().to(router::otlp::handler_receiver_otlp_traces),
                    ),
            )
            // Treat every other url as receiver endpoint
            .default_service(web::get().to(router::handler_receiver))
            // Set metrics payload limit to 100MB
            .app_data(web::PayloadConfig::default().limit(100 * 2 << 20))
    })
    .bind(format!("0.0.0.0:{}", port))?
    .run()
    .await;

    match result {
        Ok(result) => Ok(result),
        Err(e) => {
            error!("server error: {}", e);
            Err(e)
        }
    }
}
