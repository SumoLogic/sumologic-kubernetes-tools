#[macro_use]
extern crate json_str;

use std::collections::HashMap;
use std::sync::Mutex;

use actix_web::web;

use chrono::Duration;
use clap::Parser;
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

#[derive(Parser)]
#[command(
    name = "Sumo Logic Mock",
    author = "Sumo Logic <collection@sumologic.com>",
    version = "0.0",
    about = "Sumo Logic Mock can be used for testing performance or functionality of kubernetes collection without sending data to sumologic"
)]
struct Cli {
    #[arg(short, long, default_value_t = 3000, help = "Port to listen on")]
    port: u16,

    #[arg(
        short='l',
        long,
        default_value_t = String::from("localhost"), 
        help="Hostname reported as the receiver. For kubernetes it will be '<service name>.<namespace>'"
    )]
    hostname: String,

    #[arg(
        short = 'r',
        long = "print-logs",
        default_value_t = false,
        help = "Use to print received logs on stdout"
    )]
    print_logs: bool,

    #[arg(
        short = 'm',
        long = "print-metrics",
        default_value_t = false,
        help = "Use to print received metrics on stdout"
    )]
    print_metrics: bool,

    #[arg(
        short = 's',
        long = "print-spans",
        default_value_t = false,
        help = "Use to print received spans on stdout"
    )]
    print_spans: bool,

    #[arg(
        long = "print-headers",
        default_value_t = false,
        help = "Use to print received requests' headers"
    )]
    print_headers: bool,

    #[arg(
        long = "store-logs",
        default_value_t = false,
        help = "Use to store log data which can then be queried via /logs/* endpoints"
    )]
    store_logs: bool,

    #[arg(
        long = "store-metrics",
        default_value_t = false,
        help = "Use to store metrics which will then be returned via /metrics-samples"
    )]
    store_metrics: bool,

    #[arg(
        long = "store-traces",
        default_value_t = false,
        help = "Use to store traces which can then be queried via /logs/* endpoints"
    )]
    store_traces: bool,

    #[arg(
        short = 'a',
        long = "drop-rate",
        default_value_t = 0,
        help = "Use to specify packet drop rate. This is number from 0 (do not drop) to 100 (drop all)."
    )]
    drop_rate: i64,

    #[arg(
        short = 'd',
        long = "delay-time",
        default_value_t = 0,
        help = "Use to specify delay time. It mocks request processing time in milliseconds."
    )]
    delay_time: u64,
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    simple_logger::init_with_level(log::Level::Debug).unwrap();

    let cli = Cli::parse();

    let opts = Options {
        print: options::Print {
            logs: cli.print_logs,
            headers: cli.print_headers,
            metrics: cli.print_metrics,
            spans: cli.print_spans,
        },
        drop_rate: cli.drop_rate,
        delay_time: stime::Duration::from_millis(cli.delay_time),
        store_traces: cli.store_traces,
        store_metrics: cli.store_metrics,
        store_logs: cli.store_logs,
    };

    run_app(cli.hostname, cli.port, opts).await
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

    info!("Sumo Logic Mock is listening on 0.0.0.0:{}!", port);
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
                "/spans-list",
                web::get().to(router::traces_data::handler_get_spans),
            )
            .route(
                "/traces-list",
                web::get().to(router::traces_data::handler_get_traces),
            )
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
                    )
                    .route(
                        "/otCollectors/metadata",
                        web::post().to(router::api::v1::handler_collector_metadata),
                    )
                    .route(
                        "/collector/logs",
                        web::post().to(router::otlp::handler_receiver_otlp_logs),
                    )
                    .route(
                        "/collector/metrics",
                        web::post().to(router::otlp::handler_receiver_otlp_metrics),
                    )
                    .route(
                        "/collector/traces",
                        web::post().to(router::otlp::handler_receiver_otlp_traces),
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
