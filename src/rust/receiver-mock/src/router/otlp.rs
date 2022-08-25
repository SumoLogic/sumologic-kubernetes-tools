use std::collections::HashMap;
use std::io::Cursor;
use std::iter::FromIterator;

use crate::metadata::Metadata;
use crate::metrics::MetricsHandleResult;
use crate::options;
use crate::router::*;
use actix_web::{web, HttpRequest, HttpResponse, Responder};
use log::debug;
use log::warn;
use opentelemetry_proto::tonic::common::v1 as commonv1;
use opentelemetry_proto::tonic::logs::v1 as logsv1;
use opentelemetry_proto::tonic::metrics::v1 as metricsv1;
use opentelemetry_proto::tonic::trace::v1 as tracev1;
use prost::Message;

const OTLP_PROTOBUF_FORMAT_CONTENT_TYPE: &str = "application/x-protobuf";

// TODO: consider moving the whole module to a separate directory

pub async fn handler_receiver_otlp_logs(
    req: HttpRequest,
    body: web::Bytes,
    app_state: web::Data<AppState>,
    opts: web::Data<options::Options>,
) -> impl Responder {
    let remote_address = get_address(&req);
    let content_type = match get_content_type(&req) {
        Ok(x) => x,
        Err(e) => return HttpResponse::BadRequest().body(e.to_string()),
    };

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
                        debug!("log => {}", line);
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
                    Some(value) => anyvalue_to_string(value),
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
            Some(body) => anyvalue_to_string(&body),
            None => String::new(),
        })
        .collect()
}

fn anyvalue_to_string(anyvalue: &commonv1::AnyValue) -> String {
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

pub async fn handler_receiver_otlp_metrics(
    req: HttpRequest,
    body: web::Bytes,
    app_state: web::Data<AppState>,
    opts: web::Data<options::Options>,
) -> impl Responder {
    let remote_address = get_address(&req);
    let content_type = match get_content_type(&req) {
        Ok(x) => x,
        Err(e) => return HttpResponse::BadRequest().body(e.to_string()),
    };

    if let Some(response) = try_dropping_data(&opts, &content_type) {
        return response;
    }

    match content_type.as_str() {
        OTLP_PROTOBUF_FORMAT_CONTENT_TYPE => {
            let metrics_data: metricsv1::MetricsData = match metricsv1::MetricsData::decode(&mut Cursor::new(body)) {
                Ok(data) => data,
                Err(_) => return HttpResponse::BadRequest().body("Unable to parse body"),
            };
            let mut result = MetricsHandleResult::new();

            // TODO: Consider giving it some basic capacity to avoid too many allocations.
            let mut samples = vec![];
            for resource_metrics in metrics_data.resource_metrics {
                if resource_metrics.resource.is_none() {
                    warn!("resource is none for resource metrics");
                    continue;
                }

                let resource_attributes = &resource_metrics.resource.unwrap().attributes;
                for instrumentation_lib_metrics in resource_metrics.instrumentation_library_metrics {
                    for metric in instrumentation_lib_metrics.metrics {
                        let metric_sample_vec = sample::otlp_metric_to_samples(&metric, resource_attributes);

                        if opts.print.metrics {
                            for m in &metric_sample_vec {
                                debug!("metrics => {:?}", m);
                            }
                        }
                        if opts.store_metrics {
                            samples.extend(metric_sample_vec);
                        }

                        result.handle_metric(metric.name);
                        result.handle_ip(remote_address);
                    }
                }
            }

            if opts.store_metrics {
                result.metrics_samples = samples.into_iter().collect();
            }

            app_state.add_metrics_result(result, &opts);
        }
        &_ => {
            return get_invalid_header_response(&content_type);
        }
    }

    HttpResponse::Ok().body("")
}

pub async fn handler_receiver_otlp_traces(
    req: HttpRequest,
    body: web::Bytes,
    app_state: web::Data<AppState>,
    opts: web::Data<options::Options>,
) -> impl Responder {
    let _remote_address = get_address(&req);
    let content_type = match get_content_type(&req) {
        Ok(x) => x,
        Err(e) => return HttpResponse::BadRequest().body(e.to_string()),
    };

    if let Some(response) = try_dropping_data(&opts, &content_type) {
        return response;
    }

    match content_type.as_str() {
        OTLP_PROTOBUF_FORMAT_CONTENT_TYPE => {
            let traces_data = match tracev1::TracesData::decode(&mut Cursor::new(body)) {
                Ok(data) => data,
                Err(_) => return HttpResponse::BadRequest().body("Unable to parse body"),
            };
            let mut result = traces::TracesHandleResult::new();

            for resource_spans in traces_data.resource_spans {
                if resource_spans.resource.is_none() {
                    warn!("resource is none for resource spans");
                    continue;
                }

                let resource_attrs = resource_spans.resource.unwrap().attributes;
                for instrumentation_lib_spans in resource_spans.instrumentation_library_spans {
                    for span in instrumentation_lib_spans.spans {
                        let storage_span = otlp_span_to_span(&span, &resource_attrs);
                        debug!("Span => {}", storage_span);

                        result.handle_span(storage_span);
                    }
                }
            }

            app_state.add_traces_result(result, &opts);
        }
        &_ => {
            return get_invalid_header_response(&content_type);
        }
    }

    HttpResponse::Ok().body("")
}

// TODO: Move this to Sample module and rename that module.
pub fn otlp_span_to_span(otlp_span: &tracev1::Span, resource_attrs: &[commonv1::KeyValue]) -> traces::Span {
    let attributes = sample::tags_to_map(&otlp_span.attributes, resource_attrs);

    traces::Span {
        name: otlp_span.name.clone(),
        id: hex::encode(&otlp_span.span_id),
        trace_id: hex::encode(&otlp_span.trace_id),
        parent_span_id: hex::encode(&otlp_span.parent_span_id),
        attributes,
    }
}

mod sample {
    use metricsv1::number_data_point;
    use metricsv1::{Gauge, Sum};
    use opentelemetry_proto::tonic::common::v1 as commonv1;
    use opentelemetry_proto::tonic::metrics::v1 as metricsv1;
    use std::collections::HashMap;

    use crate::metrics::sample::Sample;

    type Attributes = [commonv1::KeyValue];

    const NANOS_IN_MILLIS: u64 = 1_000_000;

    pub fn otlp_metric_to_samples(metric: &metricsv1::Metric, attributes: &Attributes) -> Vec<Sample> {
        if let Some(data) = &metric.data {
            match data {
                // TODO: Support all the types
                metricsv1::metric::Data::Gauge(g) => gauge_to_samples(g, &metric.name, attributes),
                metricsv1::metric::Data::Sum(s) => sum_to_samples(s, &metric.name, attributes),
                _ => todo!(),
            }
        } else {
            vec![]
        }
    }

    fn gauge_to_samples(gauge: &Gauge, name: &str, attributes: &Attributes) -> Vec<Sample> {
        gauge
            .data_points
            .iter()
            .map(|dp| number_datapoint_to_sample(dp, name, attributes))
            .collect()
    }

    fn sum_to_samples(sum: &Sum, name: &str, attributes: &Attributes) -> Vec<Sample> {
        sum.data_points
            .iter()
            .map(|dp| number_datapoint_to_sample(dp, name, attributes))
            .collect()
    }

    fn number_datapoint_to_sample(dp: &metricsv1::NumberDataPoint, name: &str, attributes: &Attributes) -> Sample {
        if let Some(val) = &dp.value {
            let labels = tags_to_map(&dp.attributes, attributes);
            let timestamp = get_number_datapoint_timestamp_millis(dp);
            match val {
                number_data_point::Value::AsDouble(x) => create_sample(name, labels, *x, timestamp),
                number_data_point::Value::AsInt(x) => create_sample(name, labels, *x as f64, timestamp),
            }
        } else {
            Sample {
                metric: String::new(),
                value: 0.0,
                labels: HashMap::new(),
                timestamp: 0,
            }
        }
    }

    fn get_number_datapoint_timestamp_millis(dp: &metricsv1::NumberDataPoint) -> u64 {
        dp.time_unix_nano / NANOS_IN_MILLIS
    }

    fn create_sample(name: &str, labels: HashMap<String, String>, value: f64, timestamp: u64) -> Sample {
        Sample {
            metric: name.to_string(),
            value,
            labels,
            timestamp,
        }
    }

    pub fn tags_to_map(attrs: &Attributes, labels: &Attributes) -> HashMap<String, String> {
        attrs
            .iter()
            .chain(labels.iter())
            .map(|kv| {
                (
                    kv.key.clone(),
                    // FIXME: An empty string is passed instead of panicking. Some custom error could be better for debug purposes.
                    super::anyvalue_to_string(kv.value.as_ref().unwrap_or(&commonv1::AnyValue {
                        value: Some(commonv1::any_value::Value::StringValue("".to_string())),
                    })),
                )
            })
            .collect()
    }

    #[cfg(test)]
    mod test {
        use std::collections::HashMap;

        use commonv1::{AnyValue, KeyValue};
        use metricsv1::{Gauge, Metric, NumberDataPoint, Sum};
        use opentelemetry_proto::tonic::common::v1 as commonv1;
        use opentelemetry_proto::tonic::metrics::v1 as metricsv1;

        use crate::metrics::sample::Sample;

        fn get_string_anyvalue(string: &str) -> AnyValue {
            AnyValue {
                value: Some(commonv1::any_value::Value::StringValue(string.to_string())),
            }
        }

        fn pairs_to_keyvalue(pairs: Vec<(&str, AnyValue)>) -> Vec<KeyValue> {
            pairs
                .into_iter()
                .map(|(k, v)| KeyValue {
                    key: k.to_string(),
                    value: Some(v),
                })
                .collect()
        }

        fn get_sample_resource_attrs() -> Vec<KeyValue> {
            pairs_to_keyvalue(vec![
                ("key1", get_string_anyvalue("value1")),
                ("key2", get_string_anyvalue("value2")),
            ])
        }

        fn get_sample_dp_attrs() -> Vec<KeyValue> {
            pairs_to_keyvalue(vec![
                ("key2", get_string_anyvalue("surprise")),
                ("key3", get_string_anyvalue("value3")),
            ])
        }

        fn get_expected_labels() -> HashMap<String, String> {
            vec![
                ("key1".to_string(), "value1".to_string()),
                ("key2".to_string(), "value2".to_string()),
                ("key3".to_string(), "value3".to_string()),
            ]
            .into_iter()
            .collect()
        }

        #[test]
        fn otlp_format_tags_to_string_test() {
            let attrs = get_sample_dp_attrs();
            let labels = get_sample_resource_attrs();

            let expected = get_expected_labels();

            assert_eq!(super::tags_to_map(&attrs, &labels), expected);
        }

        fn get_sample_number_dp(value: i64, timestamp_ms: u64) -> NumberDataPoint {
            NumberDataPoint {
                attributes: get_sample_dp_attrs(),
                start_time_unix_nano: 161078,
                time_unix_nano: timestamp_ms * super::NANOS_IN_MILLIS,
                exemplars: vec![],
                flags: 0,
                value: Some(metricsv1::number_data_point::Value::AsInt(value)),
            }
        }

        fn get_sample_sample(name: &str, value: f64, timestamp: u64) -> Sample {
            Sample {
                metric: name.to_string(),
                value: value,
                labels: get_expected_labels(),
                timestamp,
            }
        }

        #[test]
        fn otlp_format_number_datapoint_to_sample_test() {
            let name = "metr";
            let attrs = get_sample_resource_attrs();

            let dp = get_sample_number_dp(7312, 2042005);
            let expected = get_sample_sample(name, 7312_f64, 2042005);

            assert_eq!(super::number_datapoint_to_sample(&dp, name, &attrs), expected)
        }

        pub fn get_sample_gauge() -> Gauge {
            Gauge {
                data_points: vec![get_sample_number_dp(78, 1400), get_sample_number_dp(500, 45000)],
            }
        }

        fn get_sample_sum() -> Sum {
            Sum {
                data_points: vec![get_sample_number_dp(78, 1400), get_sample_number_dp(500, 45000)],
                aggregation_temporality: 0,
                is_monotonic: true,
            }
        }

        fn get_sample_metric(name: &str, data: metricsv1::metric::Data) -> Metric {
            Metric {
                name: name.to_string(),
                description: String::new(),
                unit: String::new(),
                data: Some(data),
            }
        }

        #[test]
        fn otlp_format_gauge_to_string_test() {
            let gauge = metricsv1::metric::Data::Gauge(get_sample_gauge());
            let metric = get_sample_metric("eguag", gauge);
            let attrs = get_sample_resource_attrs();

            assert_eq!(
                super::otlp_metric_to_samples(&metric, &attrs),
                vec![
                    get_sample_sample("eguag", 78_f64, 1400),
                    get_sample_sample("eguag", 500_f64, 45000),
                ],
            )
        }

        #[test]
        fn otlp_format_sum_to_string_test() {
            let sum = metricsv1::metric::Data::Sum(get_sample_sum());
            let metric = get_sample_metric("mus", sum);
            let attrs = get_sample_resource_attrs();

            assert_eq!(
                super::otlp_metric_to_samples(&metric, &attrs),
                vec![
                    get_sample_sample("mus", 78_f64, 1400),
                    get_sample_sample("mus", 500_f64, 45000),
                ],
            )
        }
    }
}
#[cfg(test)]
mod test {
    use crate::metrics::sample::Sample;
    use crate::router::otlp::*;
    use actix_http::body::{BoxBody, MessageBody};
    use actix_web::test as actix_test;
    use actix_web::{web, App};
    use bytes::Bytes;
    use opentelemetry_proto::tonic::metrics::v1::{InstrumentationLibraryMetrics, Metric, ResourceMetrics};
    use opentelemetry_proto::tonic::{
        common::v1::{any_value::Value, AnyValue, InstrumentationLibrary, KeyValue},
        logs::v1::{InstrumentationLibraryLogs, LogRecord},
        resource::v1::Resource,
    };

    const NANOS_IN_MILLIS: u64 = 1_000_000;

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

    fn get_sample_resource() -> Resource {
        Resource {
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
        }
    }

    fn get_sample_instr_library() -> InstrumentationLibrary {
        InstrumentationLibrary {
            name: "the best library".to_string(),
            version: "v2.1.5".to_string(),
        }
    }

    fn get_sample_logs_data() -> logsv1::LogsData {
        let resource = get_sample_resource();

        let instr = vec![InstrumentationLibraryLogs {
            instrumentation_library: Some(get_sample_instr_library()),
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

    fn get_sample_logs_request_body() -> impl Into<web::Bytes> {
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

        {
            // Test logs route.
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

        {
            // Test metrics route.
            let request = actix_test::TestRequest::post()
                .uri("/v1/metrics")
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
    }

    #[actix_rt::test]
    async fn otlp_unrelated_content_type_test() {
        let content_type = "unrelated/type";
        let opts = get_default_options();
        let mut app = actix_test::init_service(
            App::new()
                .app_data(web::Data::new(opts.clone()))
                .app_data(get_default_app_data())
                .service(
                    web::scope("/v1")
                        .route("/logs", web::post().to(handler_receiver_otlp_logs))
                        .route("/metrics", web::post().to(handler_receiver_otlp_metrics)),
                )
                .default_service(web::get().to(handler_receiver)),
        )
        .await;
        {
            // Test logs route.
            let request = actix_test::TestRequest::post()
                .uri("/v1/logs")
                .insert_header(("Content-Type", content_type))
                .to_request();

            let response = actix_test::call_service(&mut app, request).await;
            assert_eq!(response.status(), StatusCode::BAD_REQUEST);
        }

        {
            // Test metrics route.
            let request = actix_test::TestRequest::post()
                .uri("/v1/metrics")
                .insert_header(("Content-Type", content_type))
                .to_request();

            let response = actix_test::call_service(&mut app, request).await;
            assert_eq!(response.status(), StatusCode::BAD_REQUEST);
        }
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
                .set_payload(get_sample_logs_request_body())
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

    fn get_string_anyvalue(string: &str) -> AnyValue {
        AnyValue {
            value: Some(commonv1::any_value::Value::StringValue(string.to_string())),
        }
    }

    fn pairs_to_keyvalue(pairs: Vec<(&str, AnyValue)>) -> Vec<KeyValue> {
        pairs
            .into_iter()
            .map(|(k, v)| KeyValue {
                key: k.to_string(),
                value: Some(v),
            })
            .collect()
    }

    fn get_sample_dp_attrs() -> Vec<KeyValue> {
        pairs_to_keyvalue(vec![
            ("key2", get_string_anyvalue("surprise")),
            ("key3", get_string_anyvalue("value3")),
        ])
    }

    fn get_sample_number_dp(value: i64, timestamp_ms: u64) -> metricsv1::NumberDataPoint {
        metricsv1::NumberDataPoint {
            attributes: get_sample_dp_attrs(),
            start_time_unix_nano: 161078,
            time_unix_nano: timestamp_ms * NANOS_IN_MILLIS,
            exemplars: vec![],
            flags: 0,
            value: Some(metricsv1::number_data_point::Value::AsInt(value)),
        }
    }

    pub fn get_sample_gauge() -> metricsv1::Gauge {
        metricsv1::Gauge {
            data_points: vec![get_sample_number_dp(78, 1400), get_sample_number_dp(500, 45000)],
        }
    }

    fn get_sample_metric(name: &str) -> Metric {
        Metric {
            name: name.to_string(),
            description: "a test metric".to_string(),
            unit: "petaweber".to_string(),
            data: Some(metricsv1::metric::Data::Gauge(get_sample_gauge())),
        }
    }

    fn get_sample_metrics_data() -> metricsv1::MetricsData {
        let resource = get_sample_resource();
        let instr = vec![InstrumentationLibraryMetrics {
            instrumentation_library: Some(get_sample_instr_library()),
            metrics: vec![get_sample_metric("length"), get_sample_metric("breath")],
            schema_url: "".to_string(),
        }];
        let resource_metrics_1 = ResourceMetrics {
            resource: Some(resource),
            instrumentation_library_metrics: instr,
            schema_url: "".to_string(),
        };
        let resource_metrics_2 = resource_metrics_1.clone();

        metricsv1::MetricsData {
            resource_metrics: vec![resource_metrics_1, resource_metrics_2],
        }
    }

    fn get_sample_metrics_request_body() -> impl Into<web::Bytes> {
        let metrics = get_sample_metrics_data();
        metrics.encode_to_vec()
    }

    #[actix_rt::test]
    async fn otlp_metrics_store_test() {
        let opts = get_default_options();
        let mut app = actix_test::init_service(
            App::new()
                .app_data(web::Data::new(opts.clone()))
                .app_data(get_default_app_data())
                .service(web::scope("/v1").route("/metrics", web::post().to(handler_receiver_otlp_metrics)))
                .route(
                    "/metrics-list",
                    web::get().to(metrics_data::handler_metrics_list),
                )
                .default_service(web::get().to(handler_receiver)),
        )
        .await;

        {
            let request = actix_test::TestRequest::post()
                .uri("/v1/metrics")
                .insert_header(("Content-Type", OTLP_PROTOBUF_FORMAT_CONTENT_TYPE))
                .set_payload(get_sample_metrics_request_body())
                .to_request();

            let response = actix_test::call_service(&mut app, request).await;
            assert_eq!(response.status(), StatusCode::OK);
        }

        {
            let request = actix_test::TestRequest::get().uri("/metrics-list").to_request();

            let response = actix_test::call_service(&mut app, request).await;
            assert_eq!(response.status(), StatusCode::OK);

            let body = actix_test::read_body(response).await;

            assert!(body.eq(&Bytes::from("length: 2\nbreath: 2\n")) || body.eq(&Bytes::from("breath: 2\nlength: 2\n")));
        }
    }

    #[actix_rt::test]
    async fn otlp_metrics_store_samples_test() {
        let opts = get_default_options();
        let mut app = actix_test::init_service(
            App::new()
                .app_data(web::Data::new(opts.clone()))
                .app_data(get_default_app_data())
                .service(web::scope("/v1").route("/metrics", web::post().to(handler_receiver_otlp_metrics)))
                .route(
                    "/metrics-samples",
                    web::get().to(metrics_data::handler_metrics_samples),
                )
                .default_service(web::get().to(handler_receiver)),
        )
        .await;

        {
            let request = actix_test::TestRequest::post()
                .uri("/v1/metrics")
                .insert_header(("Content-Type", OTLP_PROTOBUF_FORMAT_CONTENT_TYPE))
                .set_payload(get_sample_metrics_request_body())
                .to_request();

            let response = actix_test::call_service(&mut app, request).await;
            assert_eq!(response.status(), StatusCode::OK);
        }

        {
            let request = actix_test::TestRequest::get()
                .uri("/metrics-samples")
                .set_payload(get_sample_metrics_request_body())
                .to_request();
            let response = actix_test::call_service(&mut app, request).await;

            assert_eq!(response.status(), StatusCode::OK);

            let result: Vec<Sample> = actix_test::read_body_json(response).await;
            assert_eq!(result.len(), 2);
        }
    }
}
