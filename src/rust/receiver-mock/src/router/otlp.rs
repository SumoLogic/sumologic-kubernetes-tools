use std::collections::HashMap;
use std::io::Cursor;
use std::iter::FromIterator;

use crate::metadata::Metadata;
use crate::options;
use crate::router::*;
use actix_web::{web, HttpRequest, HttpResponse, Responder};
use opentelemetry_proto::tonic::common::v1::{self as commonv1};
use opentelemetry_proto::tonic::logs::v1 as logsv1;
use prost::Message;

const OTLP_PROTOBUF_FORMAT_CONTENT_TYPE: &str = "application/x-protobuf";

pub async fn handler_receiver_otlp_logs(
    req: HttpRequest,
    body: web::Bytes,
    app_state: web::Data<AppState>,
    opts: web::Data<options::Options>,
) -> impl Responder {
    let remote_address = get_address(&req);
    let content_type = get_content_type(&req);

    if let Some(response) = try_dropping_data(&opts, &content_type) {
        return response;
    }

    match content_type.as_str() {
        OTLP_PROTOBUF_FORMAT_CONTENT_TYPE => {
            let log_data: logsv1::LogsData = match logsv1::LogsData::decode(&mut Cursor::new(body)) {
                Ok(data) => data,
                Err(_) => return HttpResponse::BadRequest().body("Unable to parse body"),
            };
            for resource_logs in log_data.resource_logs {
                let metadata = get_otlp_metadata_from_logs(&resource_logs);
                let lines = get_otlp_lines_from_logs(&resource_logs);

                app_state.add_log_lines(
                    lines.iter().map(|x| x.as_str()),
                    metadata,
                    remote_address,
                    &opts,
                );

                if opts.print.logs {
                    for line in lines {
                        println!("log => {}", line);
                    }
                }
            }
        }
        &_ => {
            return get_invalid_header_response(&content_type);
        }
    }

    HttpResponse::Ok().body("")
}

fn get_otlp_metadata_from_logs(resource_logs: &logsv1::ResourceLogs) -> Metadata {
    match &resource_logs.resource {
        Some(resource) => HashMap::from_iter(resource.attributes.iter().map(|kv| {
            (
                kv.key.clone(),
                match &kv.value {
                    Some(value) => otlp_anyvalue_to_string(value),
                    None => String::new(),
                },
            )
        })),
        None => Metadata::new(),
    }
}

fn get_otlp_lines_from_logs(resource_logs: &logsv1::ResourceLogs) -> Vec<String> {
    resource_logs
        .instrumentation_library_logs
        .iter()
        .map(|ill| ill.log_records.iter())
        .flatten()
        .map(|log_record| match &log_record.body {
            Some(body) => otlp_anyvalue_to_string(&body),
            None => String::new(),
        })
        .collect()
}

fn otlp_anyvalue_to_string(anyvalue: &commonv1::AnyValue) -> String {
    let value = match &anyvalue.value {
        Some(v) => v,
        None => return String::new(),
    };
    let s = match value {
        commonv1::any_value::Value::StringValue(inner) => inner.clone(),
        commonv1::any_value::Value::BoolValue(inner) => inner.to_string(),
        commonv1::any_value::Value::IntValue(inner) => inner.to_string(),
        commonv1::any_value::Value::DoubleValue(inner) => inner.to_string(),
        _ => String::new(),
    };

    return s;
}

#[cfg(test)]
mod test {
    use crate::router::otlp::*;
    use actix_http::body::{BoxBody, MessageBody};
    use actix_web::test as actix_test;
    use actix_web::{web, App};
    use opentelemetry_proto::tonic::{
        common::v1::{any_value::Value, AnyValue, InstrumentationLibrary, KeyValue},
        logs::v1::{InstrumentationLibraryLogs, LogRecord},
        resource::v1::Resource,
    };

    fn get_default_options() -> options::Options {
        options::Options {
            print: options::Print {
                logs: false,
                headers: false,
                metrics: false,
            },
            delay_time: std::time::Duration::from_secs(0),
            drop_rate: 0,
            store_metrics: true,
            store_logs: true,
        }
    }

    fn get_default_app_data() -> web::Data<AppState> {
        web::Data::new(AppState::new())
    }

    fn get_body_str(body: BoxBody) -> String {
        std::str::from_utf8(&body.try_into_bytes().unwrap())
            .unwrap()
            .to_string()
    }

    fn get_sample_log_record(body: &str) -> LogRecord {
        #[allow(deprecated)]
        LogRecord {
            time_unix_nano: 21,
            observed_time_unix_nano: 99,
            severity_number: 20000,
            severity_text: "warning".to_string(),
            body: Some(AnyValue {
                value: Some(Value::StringValue(body.to_string())),
            }),
            name: "temperature log".to_string(),
            attributes: vec![],
            dropped_attributes_count: 0,
            flags: 0b101010,
            trace_id: vec![],
            span_id: vec![],
        }
    }

    fn get_sample_logs_data() -> logsv1::LogsData {
        let resource = Resource {
            attributes: vec![
                KeyValue {
                    key: "some-key".to_string(),
                    value: Some(AnyValue {
                        value: Some(Value::StringValue("blep".to_string())),
                    }),
                },
                KeyValue {
                    key: "another-key".to_string(),
                    value: Some(AnyValue {
                        value: Some(Value::StringValue("qwerty".to_string())),
                    }),
                },
            ],
            dropped_attributes_count: 0,
        };

        let instr = vec![InstrumentationLibraryLogs {
            instrumentation_library: Some(InstrumentationLibrary {
                name: "the best library".to_string(),
                version: "v2.1.5".to_string(),
            }),
            log_records: vec![
                get_sample_log_record("warning: the temperature is too low"),
                get_sample_log_record("killing child with a fork"),
            ],
            schema_url: String::new(),
        }];

        let resource_logs_1 = logsv1::ResourceLogs {
            resource: Some(resource.clone()),
            instrumentation_library_logs: instr,
            schema_url: String::new(),
        };

        let resource_logs_2 = resource_logs_1.clone();
        logsv1::LogsData {
            resource_logs: vec![resource_logs_1, resource_logs_2],
        }
    }

    fn get_sample_request_body() -> impl Into<web::Bytes> {
        let logs = get_sample_logs_data();
        logs.encode_to_vec()
    }

    #[test]
    fn otlp_logs_get_metadata_test() {
        let logs = &get_sample_logs_data().resource_logs[0];
        let metadata = get_otlp_metadata_from_logs(logs);

        let mut expected = HashMap::new();
        expected.insert("some-key".to_string(), "blep".to_string());
        expected.insert("another-key".to_string(), "qwerty".to_string());

        assert_eq!(metadata, expected)
    }

    #[test]
    fn otlp_logs_get_lines_test() {
        let logs = &get_sample_logs_data().resource_logs[0];
        let lines = get_otlp_lines_from_logs(logs);

        let expected = vec!["warning: the temperature is too low", "killing child with a fork"];

        assert_eq!(lines, expected)
    }

    #[actix_rt::test]
    async fn otlp_logs_drop_test() {
        let mut opts = get_default_options();
        opts.drop_rate = 100;

        let mut app = actix_test::init_service(
            App::new()
                .app_data(web::Data::new(opts.clone()))
                .app_data(get_default_app_data())
                .service(web::scope("/v1").route("/logs", web::post().to(handler_receiver_otlp_logs)))
                .default_service(web::get().to(handler_receiver)),
        )
        .await;

        let request = actix_test::TestRequest::post()
            .uri("/v1/logs")
            .insert_header(("Content-Type", OTLP_PROTOBUF_FORMAT_CONTENT_TYPE))
            .to_request();

        let response = actix_test::call_service(&mut app, request).await;
        assert_eq!(response.status(), StatusCode::INTERNAL_SERVER_ERROR);
        let body = response.into_body();
        assert_eq!(
            get_body_str(body),
            format!("Dropping data for {}", OTLP_PROTOBUF_FORMAT_CONTENT_TYPE)
        );
    }

    #[actix_rt::test]
    async fn otlp_logs_unrelated_content_type_test() {
        let content_type = "unrelated/type";
        let opts = get_default_options();
        let mut app = actix_test::init_service(
            App::new()
                .app_data(web::Data::new(opts.clone()))
                .app_data(get_default_app_data())
                .service(web::scope("/v1").route("/logs", web::post().to(handler_receiver_otlp_logs)))
                .default_service(web::get().to(handler_receiver)),
        )
        .await;

        let request = actix_test::TestRequest::post()
            .uri("/v1/logs")
            .insert_header(("Content-Type", content_type))
            .to_request();

        let response = actix_test::call_service(&mut app, request).await;
        assert_eq!(response.status(), StatusCode::BAD_REQUEST);
    }

    #[actix_rt::test]
    async fn otlp_logs_store_test() {
        let opts = get_default_options();
        let mut app = actix_test::init_service(
            App::new()
                .app_data(web::Data::new(opts.clone()))
                .app_data(get_default_app_data())
                .service(web::scope("/v1").route("/logs", web::post().to(handler_receiver_otlp_logs)))
                .route("/logs/count", web::get().to(handler_logs_count))
                .default_service(web::get().to(handler_receiver)),
        )
        .await;

        {
            let request = actix_test::TestRequest::post()
                .uri("/v1/logs")
                .insert_header(("Content-Type", OTLP_PROTOBUF_FORMAT_CONTENT_TYPE))
                .set_payload(get_sample_request_body())
                .to_request();

            let response = actix_test::call_service(&mut app, request).await;
            assert_eq!(response.status(), StatusCode::OK);
        }

        // count all the logs
        {
            let req = actix_test::TestRequest::get().uri("/logs/count").to_request();
            let resp = actix_test::call_service(&mut app, req).await;

            let response_body: LogsCountResponse = actix_test::read_body_json(resp).await;

            assert_eq!(response_body.count, 4);
        }
    }
}
