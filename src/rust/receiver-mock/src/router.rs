use std::collections::{HashMap, HashSet};
use std::net::{IpAddr, Ipv4Addr};
use std::sync::{Mutex, RwLock};

use actix_http::http;
use actix_web::{http::StatusCode, web, HttpRequest, HttpResponse, Responder};
use bytes;
use rand::distributions::Alphanumeric;
use rand::{thread_rng, Rng};
use serde_derive::{Deserialize, Serialize};

use crate::logs;
use crate::metrics;
use crate::metrics::Sample;
use crate::options;
use crate::time::get_now;

pub struct AppState {
    // Mutexes are necessary to mutate data safely across threads in handlers.
    //
    pub metrics: Mutex<u64>,
    pub logs: RwLock<logs::LogRepository>,

    pub metrics_samples: Mutex<HashSet<Sample>>,
    pub metrics_list: Mutex<HashMap<String, u64>>,
    pub metrics_ip_list: Mutex<HashMap<IpAddr, u64>>,
}

impl AppState {
    pub fn new() -> Self {
        return Self {
            logs: RwLock::new(logs::LogRepository::new()),

            metrics: Mutex::new(0),
            metrics_list: Mutex::new(HashMap::new()),
            metrics_ip_list: Mutex::new(HashMap::new()),
            metrics_samples: Mutex::new(HashSet::new()),
        };
    }
}

impl AppState {
    pub fn add_metrics_result(&self, result: metrics::MetricsHandleResult, opts: &options::Options) {
        {
            let mut metrics = self.metrics.lock().unwrap();
            *metrics += result.metrics_count;
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

        if opts.store_metrics {
            // Replace old data points that represent the same data series
            // (the same metric name and labels) with new ones.
            let mut samples = self.metrics_samples.lock().unwrap();
            for s in result.metrics_samples {
                samples.replace(s);
            }
        }
    }
}

pub struct AppMetadata {
    pub url: String,
}

pub struct TerraformState {
    pub fields: Mutex<HashMap<String, String>>,
}

// Reset metrics counter
pub async fn handler_metrics_reset(app_state: web::Data<AppState>) -> impl Responder {
    *app_state.metrics.lock().unwrap() = 0;
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

// Return consumed metrics. Endpoint handled by this handler will accept URL query
// which is a key value list of label names and values that the metric should contain
// in order be returned.
// Label values can be ommitted in which case only presence of a particular label
// will be checked, not its value in the filtered sample.
// `__name__` is handled specially as it will be matched against the metric name.
//
// Exemplar usage of this endpoint:
//
// $ curl -s localhost:3000/metrics-samples\?__name__=apiserver_request_total\&cluster | jq .
// [
//     {
//       "metric": "apiserver_request_total",
//       "value": 124,
//       "labels": {
//         "prometheus_replica": "prometheus-release-test-1638873119-ku-prometheus-0",
//         "_origin": "kubernetes",
//         "component": "apiserver",
//         "service": "kubernetes",
//         "resource": "events",
//         "code": "422",
//         "instance": "172.18.0.2:6443",
//         "group": "events.k8s.io",
//         "namespace": "default",
//         "verb": "POST",
//         "scope": "resource",
//         "endpoint": "https",
//         "version": "v1",
//         "job": "apiserver",
//         "cluster": "microk8s",
//         "prometheus": "ns-test-1638873119/release-test-1638873119-ku-prometheus"
//       },
//       "timestamp": 163123123
//     }
//   ]
//
pub async fn handler_metrics_samples(
    app_state: web::Data<AppState>,
    web::Query(params): web::Query<HashMap<String, String>>,
    opts: web::Data<options::Options>,
) -> impl Responder {
    if !opts.store_metrics {
        return HttpResponse::NotImplemented().body("");
    }

    let samples = &*app_state.metrics_samples.lock().unwrap();
    let response = metrics::filter_samples(samples, params);

    HttpResponse::Ok().json(response)
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
        app_state.logs.read().unwrap().total.count,
        app_state.logs.read().unwrap().total.bytes,
    );

    {
        let metrics_ip_list = app_state.metrics_ip_list.lock().unwrap();
        if metrics_ip_list.len() > 0 {
            let mut metrics_ip_string = String::from("# TYPE receiver_mock_metrics_ip_count counter\n");
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
        let logs = app_state.logs.read().unwrap();
        if logs.ipaddr_to_stats.len() > 0 {
            let mut logs_ip_count_bytes_string = String::from("# TYPE receiver_mock_logs_bytes_ip_count counter\n");
            let mut logs_ip_count_string = String::from("# TYPE receiver_mock_logs_ip_count counter\n");

            for (ip, val) in logs.ipaddr_to_stats.iter() {
                logs_ip_count_string.push_str(&format!(
                    "receiver_mock_logs_ip_count{{ip_address=\"{}\"}} {}\n",
                    ip, val.count
                ));
                logs_ip_count_bytes_string.push_str(&format!(
                    "receiver_mock_logs_bytes_ip_count{{ip_address=\"{}\"}} {}\n",
                    ip, val.bytes
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
    struct Source {
        url: String,
    }

    #[derive(Serialize)]
    struct TerraformResponse {
        source: Source,
    }

    web::Json(TerraformResponse {
        source: Source {
            url: app_metadata.url.clone(),
        },
    })
}

pub async fn handler_terraform_fields_quota() -> impl Responder {
    #[derive(Serialize)]
    struct TerraformFieldsQuotaResponse {
        quota: u64,
        remaining: u64,
    }

    web::Json(TerraformFieldsQuotaResponse {
        quota: 200,
        remaining: 100,
    })
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
struct TerraformFieldObject {
    field_name: String,
    field_id: String,
    data_type: String,
    state: String,
}

pub async fn handler_terraform_fields(terraform_state: web::Data<TerraformState>) -> impl Responder {
    #[derive(Serialize)]
    struct TerraformFieldsResponse {
        data: Vec<TerraformFieldObject>,
    }

    let fields = terraform_state.fields.lock().unwrap();
    let res = fields.iter().map(|(id, name)| TerraformFieldObject {
        field_name: name.clone(),
        field_id: id.clone(),
        data_type: String::from("String"),
        state: String::from("Enabled"),
    });

    web::Json(TerraformFieldsResponse { data: res.collect() })
}

#[derive(Deserialize)]
pub struct TerraformFieldParams {
    field: String,
}

pub async fn handler_terraform_field(
    params: web::Path<TerraformFieldParams>,
    terraform_state: web::Data<TerraformState>,
) -> impl Responder {
    #[derive(Debug, Serialize)]
    struct TerraformFieldResponseErrorMetaField {
        id: String,
    }

    #[derive(Debug, Serialize)]
    struct TerraformFieldResponseErrorField {
        code: String,
        message: String,
        meta: TerraformFieldResponseErrorMetaField,
    }

    #[derive(Debug, Serialize)]
    struct TerraformFieldNotFoundError {
        id: String,
        errors: Vec<TerraformFieldResponseErrorField>,
    }

    let fields = terraform_state.fields.lock().unwrap();
    let id = params.field.clone();
    let res = fields.get(&id);

    match res {
        Some(name) => HttpResponse::build(StatusCode::OK).json(TerraformFieldObject {
            field_name: name.clone(),
            field_id: id,
            data_type: String::from("String"),
            state: String::from("Enabled"),
        }),

        None => HttpResponse::build(StatusCode::NOT_FOUND).json(TerraformFieldNotFoundError {
            id: String::from("QL6LR-5P7KI-RAR20"),
            errors: vec![TerraformFieldResponseErrorField {
                code: String::from("field:doesnt_exist"),
                message: String::from("Field with the given id doesn't exist"),
                meta: TerraformFieldResponseErrorMetaField { id: id },
            }],
        }),
    }
}

#[derive(Deserialize, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct TerraformFieldCreateRequest {
    field_name: String,
}

// This handler needs to keep the information about created fields because
// sumologic terraform provider looks up those fields after creation and if it
// cannot find it (via a GET request to /api/v1/fields) then it fails.
//
// ref: https://github.com/SumoLogic/terraform-provider-sumologic/blob/3275ce043e08873a8b1b843d4a5d473619044b4e/sumologic%2Fresource_sumologic_field.go#L95-L109
// ref: https://github.com/SumoLogic/terraform-provider-sumologic/blob/3275ce043e08873a8b1b843d4a5d473619044b4e/sumologic%2Fresource_sumologic_field.go#L60
//
pub async fn handler_terraform_fields_create(
    req: web::Json<TerraformFieldCreateRequest>,
    terraform_state: web::Data<TerraformState>,
) -> impl Responder {
    let mut fields = terraform_state.fields.lock().unwrap();
    let id: String = thread_rng()
        .sample_iter(Alphanumeric)
        .take(16)
        .map(char::from)
        .collect();
    let requested_name = req.field_name.clone();
    let exists = fields.iter().find_map(
        |(id, name)| {
            if name == &requested_name {
                Some(id)
            } else {
                None
            }
        },
    );

    match exists {
        // Field with given ID already existed
        Some(_) => HttpResponse::from(json_str!({
            id: "E40YU-CU3Q7-RQDMO",
            errors: [
                {
                    code: "field:already_exists",
                    message: "Field with the given name already exists"
                }
            ]
        })),

        // New field can be inserted
        None => {
            let res = fields.insert(id.clone(), requested_name.clone());
            match res {
                None => HttpResponse::build(StatusCode::OK).json(TerraformFieldObject {
                    field_name: requested_name.clone(),
                    field_id: id,
                    data_type: String::from("String"),
                    state: String::from("Enabled"),
                }),

                // This theoretically shouldn't happen but just in case handle this
                Some(_) => HttpResponse::from(json_str!({
                    id: "E40YU-CU3Q7-RQDMO",
                    errors: [
                        {
                            code: "field:already_exists",
                            message: "Field with the given name already exists"
                        }
                    ]

                })),
            }
        }
    }
}

pub async fn handler_receiver(
    req: HttpRequest,
    body: bytes::Bytes,
    app_state: web::Data<AppState>,
    opts: web::Data<options::Options>,
) -> impl Responder {
    // Don't fail when we can't read remote address.
    // Default to localhost and just ingest what was sent.
    let localhost: std::net::SocketAddr = std::net::SocketAddr::new(IpAddr::V4(Ipv4Addr::new(127, 0, 0, 1)), 0);
    let remote_address = req.peer_addr().unwrap_or(localhost).ip();

    // actix automatically decompresses body for us.
    let string_body = String::from_utf8(body.to_vec()).unwrap();
    let lines = string_body.trim().lines();

    let headers = req.headers();
    let empty_header = http::HeaderValue::from_str("").unwrap();
    let content_type = headers.get("content-type").unwrap_or(&empty_header).to_str().unwrap();

    let mut rng = rand::thread_rng();
    let number: i64 = rng.gen_range(0..100);
    if number < opts.drop_rate {
        println!("Dropping data for {}", content_type);
        return HttpResponse::InternalServerError();
    }

    match content_type {
        // Metrics in carbon2 format
        "application/vnd.sumologic.carbon2" => {
            let result = metrics::handle_carbon2(lines, remote_address, opts.print);
            app_state.add_metrics_result(result, opts.get_ref());
        }

        // Metrics in graphite format
        "application/vnd.sumologic.graphite" => {
            let result = metrics::handle_graphite(lines, remote_address, opts.print);
            app_state.add_metrics_result(result, opts.get_ref());
        }

        // Metrics in prometheus format
        "application/vnd.sumologic.prometheus" => {
            let result = metrics::handle_prometheus(lines, remote_address, opts.get_ref());
            app_state.add_metrics_result(result, opts.get_ref());
        }

        // Logs & events
        "application/x-www-form-urlencoded" => {
            {
                let mut log_repository = app_state.logs.write().unwrap();
                for line in lines.clone() {
                    log_repository.add_log_message(line.to_string(), remote_address)
                }
            }
            if opts.print.logs {
                for line in lines.clone() {
                    println!("log => {}", line);
                }
            }
        }

        &_ => {
            println!("invalid header value");
        }
    }

    HttpResponse::Ok()
}

pub fn print_request_headers(
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
    interval: chrono::Duration,
    app_state: web::Data<AppState>,
) -> timer::Guard {
    let mut p_metrics: u64 = 0;
    let mut p_logs: u64 = 0;
    let mut p_logs_bytes: u64 = 0;
    let mut ts = get_now();

    t.schedule_repeating(interval, move || {
        let now = get_now();
        let metrics = app_state.metrics.lock().unwrap();
        let logs = app_state.logs.read().unwrap();

        // TODO: make this print metrics per minute (as DPM) and logs
        // per second, regardless of used interval
        // ref: https://github.com/SumoLogic/sumologic-kubernetes-tools/issues/57
        println!(
            "{} Metrics: {:10.} Logs: {:10.}; {:6.6} MB/s",
            now,
            *metrics - p_metrics,
            logs.total.count - p_logs,
            ((logs.total.bytes - p_logs_bytes) as f64) / ((now - ts) as f64) / (1e6 as f64)
        );

        ts = now;
        p_metrics = *metrics;
        p_logs = logs.total.count;
        p_logs_bytes = logs.total.bytes;
    })
}

#[cfg(test)]
mod tests_terraform {
    use super::*;
    use actix_rt;
    use actix_web::{test, web, App};
    use futures_util::stream::TryStreamExt;

    #[actix_rt::test]
    async fn test_handler_terraform() {
        let app_metadata = web::Data::new(AppMetadata {
            url: String::from("http://hostname:3000/terraform"),
        });
        let mut app = test::init_service(
            App::new().service(
                web::scope("/terraform")
                    .app_data(app_metadata)
                    .default_service(web::get().to(handler_terraform)),
            ),
        )
        .await;

        {
            let req = test::TestRequest::get().uri("/terraform").to_request();
            let mut resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let bytes = test::load_stream(resp.take_body().into_stream()).await;
            assert_eq!(
                bytes.unwrap(),
                json_str!({
                    source: {
                      url: "http://hostname:3000/terraform"
                    }
                })
            );
        }
        {
            let req = test::TestRequest::get()
                .uri("/terraform/api/v1/collectors/0/sources/0")
                .to_request();
            let mut resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let bytes = test::load_stream(resp.take_body().into_stream()).await;
            assert_eq!(
                bytes.unwrap(),
                json_str!({
                    source: {
                      url: "http://hostname:3000/terraform"
                    }
                })
            );
        }
        {
            let req = test::TestRequest::get().uri("/different_route").to_request();
            let mut resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 404);

            let bytes = test::load_stream(resp.take_body().into_stream()).await;
            assert_eq!(bytes.unwrap(), "");
        }
    }

    #[actix_rt::test]
    async fn test_handler_terraform_fields_quota() {
        let app_metadata = web::Data::new(AppMetadata {
            url: String::from("http://hostname:3000/receiver"),
        });

        let web_data_app_state = web::Data::new(AppState::new());

        let mut app = test::init_service(
            App::new()
                .app_data(web_data_app_state.clone()) // Mutable shared state
                .service(
                    web::scope("/terraform")
                        .app_data(app_metadata.clone())
                        .route(
                            "/api/v1/fields/quota",
                            web::get().to(handler_terraform_fields_quota),
                        )
                        .default_service(web::get().to(handler_terraform)),
                ),
        )
        .await;

        {
            let req = test::TestRequest::get()
                .uri("/terraform/api/v1/fields/quota")
                .to_request();
            let mut resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let bytes = test::load_stream(resp.take_body().into_stream()).await;

            assert_eq!(bytes.unwrap(), r#"{"quota":200,"remaining":100}"#);
        }
    }

    #[actix_rt::test]
    async fn test_handler_terraform_fields() {
        let app_metadata = web::Data::new(AppMetadata {
            url: String::from("http://hostname:3000/receiver"),
        });

        let web_data_app_state = web::Data::new(AppState::new());

        let terraform_state = web::Data::new(TerraformState {
            fields: Mutex::new(HashMap::new()),
        });

        let mut app = test::init_service(
            App::new()
                .app_data(web_data_app_state.clone()) // Mutable shared state
                .service(
                    web::scope("/terraform")
                        .app_data(app_metadata.clone())
                        .app_data(terraform_state.clone())
                        .route("/api/v1/fields/{field}", web::get().to(handler_terraform_field))
                        .route(
                            "/api/v1/fields",
                            web::post().to(handler_terraform_fields_create),
                        )
                        .route("/api/v1/fields", web::get().to(handler_terraform_fields))
                        .default_service(web::get().to(handler_terraform)),
                ),
        )
        .await;

        {
            let req = test::TestRequest::get().uri("/terraform/api/v1/fields").to_request();
            let mut resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let bytes = test::load_stream(resp.take_body().into_stream()).await;

            assert_eq!(
                bytes.unwrap(),
                json_str!({
                    data: []
                })
            );
        }

        {
            let req = test::TestRequest::get()
                .uri("/terraform/api/v1/fields/dummyID123")
                .to_request();
            let mut resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 404);

            let bytes = test::load_stream(resp.take_body().into_stream()).await;

            assert_eq!(
                bytes.unwrap(),
                json_str!({
                    id: "QL6LR-5P7KI-RAR20",
                    errors: [
                        {
                            code: "field:doesnt_exist",
                            message: "Field with the given id doesn't exist",
                            meta: {
                                id: "dummyID123"
                            }
                        }
                    ]
                })
            );
        }

        {
            let req = test::TestRequest::get().uri("/terraform/api/v1/fields").to_request();
            let mut resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let bytes = test::load_stream(resp.take_body().into_stream()).await;

            assert_eq!(
                bytes.unwrap(),
                json_str!({
                    data: []
                })
            );
        }

        {
            let req = test::TestRequest::post()
                .uri("/terraform/api/v1/fields")
                .set_json(&TerraformFieldCreateRequest {
                    field_name: String::from("dummyID123"),
                })
                .to_request();
            let mut resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let bytes = test::load_stream(resp.take_body().into_stream()).await;

            // Probably add more checks about the returned body: generated random
            // IDs are a bit problematic here.
            assert_ne!(
                bytes.unwrap(),
                json_str!({
                    data: []
                })
            );
        }

        {
            let req = test::TestRequest::get().uri("/terraform/api/v1/fields").to_request();
            let mut resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let bytes = test::load_stream(resp.take_body().into_stream()).await;

            // Probably add more checks about the returned body: generated random
            // IDs are a bit problematic here.
            assert_ne!(
                bytes.unwrap(),
                json_str!({
                    data: []
                })
            );
        }
    }
}

#[cfg(test)]
mod tests_metrics {
    use super::*;
    use actix_rt;
    use actix_web::{test, web, App};
    use futures_util::stream::TryStreamExt;
    use std::array::IntoIter;
    use std::iter::FromIterator;

    #[actix_rt::test]
    async fn test_handler_metrics_reset() {
        let mut metrics_list = HashMap::new();
        metrics_list.insert(String::from("mem_active"), 1000);
        metrics_list.insert(String::from("mem_free"), 2000);

        let app_state = AppState::new();
        *app_state.metrics.lock().unwrap() = 3000;
        *app_state.metrics_list.lock().unwrap() = metrics_list;
        let web_data_app_state = web::Data::new(app_state);

        let mut app = test::init_service(
            App::new()
                .app_data(web_data_app_state.clone()) // Mutable shared state
                .route("/metrics-reset", web::post().to(handler_metrics_reset))
                .route("/metrics", web::get().to(handler_metrics)),
        )
        .await;

        {
            let req = test::TestRequest::get().uri("/metrics").to_request();
            let mut resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let bytes = test::load_stream(resp.take_body().into_stream()).await;
            assert_eq!(
                bytes.unwrap(),
                r#"# TYPE receiver_mock_metrics_count counter
receiver_mock_metrics_count 3000
# TYPE receiver_mock_logs_count counter
receiver_mock_logs_count 0
# TYPE receiver_mock_logs_bytes_count counter
receiver_mock_logs_bytes_count 0
"#
            );
        }
        {
            let req = test::TestRequest::post().uri("/metrics-reset").to_request();
            let mut resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let bytes = test::load_stream(resp.take_body().into_stream()).await;
            assert_eq!(bytes.unwrap(), "All counters reset successfully");
        }
        {
            let req = test::TestRequest::get().uri("/metrics").to_request();
            let mut resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let bytes = test::load_stream(resp.take_body().into_stream()).await;
            assert_eq!(
                bytes.unwrap(),
                r#"# TYPE receiver_mock_metrics_count counter
receiver_mock_metrics_count 0
# TYPE receiver_mock_logs_count counter
receiver_mock_logs_count 0
# TYPE receiver_mock_logs_bytes_count counter
receiver_mock_logs_bytes_count 0
"#
            );
        }
    }

    #[actix_rt::test]
    async fn test_handler_metrics_list() {
        let mut metrics_list = HashMap::new();
        metrics_list.insert(String::from("mem_free"), 2000);

        let app_state = AppState::new();
        *app_state.metrics.lock().unwrap() = 2000;
        *app_state.metrics_list.lock().unwrap() = metrics_list;
        let web_data_app_state = web::Data::new(app_state);

        let mut app = test::init_service(
            App::new()
                .app_data(web_data_app_state.clone()) // Mutable shared state
                .route("/metrics-list", web::get().to(handler_metrics_list))
                .route("/metrics", web::get().to(handler_metrics)),
        )
        .await;

        {
            let req = test::TestRequest::get().uri("/metrics-list").to_request();
            let mut resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let bytes = test::load_stream(resp.take_body().into_stream()).await;

            // Testing multiple metric names being returned properly would be nice
            // but since we're using a hashmap we'd need to sort the lines from the byte
            // buffer that we receive and ain't that easy (but definitely doable).
            assert_eq!(bytes.unwrap(), "mem_free: 2000\n");
        }
    }

    #[actix_rt::test]
    async fn test_handler_metrics_storage() {
        let web_data_app_state = web::Data::new(AppState::new());
        let opts = options::Options {
            print: options::Print {
                logs: false,
                headers: false,
                metrics: false,
            },
            drop_rate: 0,
            store_metrics: true,
        };

        let mut app = test::init_service(
            actix_web::App::new()
                .data(opts.clone())
                .app_data(web_data_app_state.clone()) // Mutable shared state
                .route("/metrics-samples", web::get().to(handler_metrics_samples))
                .default_service(web::get().to(handler_receiver)),
        )
        .await;

        {
            let req = test::TestRequest::post().uri("/")
            .set_payload(r#"apiserver_request_total{cluster="microk8s",mock="yes",code="200",component="apiserver",endpoint="https",group="events.k8s.io",job="apiserver"} 123.12 1638873379541"#)
            .header("Content-Type", "application/vnd.sumologic.prometheus")
            .to_request();

            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);
        }
        {
            let req = test::TestRequest::get().uri("/metrics-samples").to_request();

            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let result: Vec<Sample> = test::read_body_json(resp).await;
            assert_eq!(result.len(), 1);
            assert_eq!(result[0].metric, "apiserver_request_total");
            assert_eq!(result[0].value, 123.12);
            assert_eq!(result[0].timestamp, 1638873379541);
            assert_eq!(
                result[0].labels,
                // ref: https://stackoverflow.com/a/27582993
                HashMap::<String, String>::from_iter(IntoIter::new([
                    ("mock".to_owned(), "yes".to_owned()),
                    ("group".to_owned(), "events.k8s.io".to_owned()),
                    ("code".to_owned(), "200".to_owned()),
                    ("job".to_owned(), "apiserver".to_owned()),
                    ("cluster".to_owned(), "microk8s".to_owned()),
                    ("component".to_owned(), "apiserver".to_owned()),
                    ("endpoint".to_owned(), "https".to_owned()),
                ]))
            );
        }
        {
            // Another request with a different time series (different labels set)
            // should produce a different/new time series
            let req = test::TestRequest::post().uri("/")
            .set_payload(r#"apiserver_request_total{cluster="microk8s",code="200",component="apiserver",endpoint="https",group="events.k8s.io",job="apiserver",namespace="default",resource="events"} 128.12 1638873379541"#)
            .header("Content-Type", "application/vnd.sumologic.prometheus")
            .to_request();

            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);
        }
        {
            // Let's check those by adding URL query params
            // This time series has a namespace & resources labels while the other
            // one doesn't.
            let req = test::TestRequest::get()
                .uri("/metrics-samples?resource=events")
                .to_request();

            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let result: Vec<Sample> = test::read_body_json(resp).await;
            assert_eq!(result.len(), 1);
            assert_eq!(result[0].metric, "apiserver_request_total");
            assert_eq!(result[0].value, 128.12);
            assert_eq!(result[0].timestamp, 1638873379541);
            assert_eq!(
                result[0].labels,
                HashMap::<String, String>::from_iter(IntoIter::new([
                    ("cluster".to_owned(), "microk8s".to_owned()),
                    ("code".to_owned(), "200".to_owned()),
                    ("component".to_owned(), "apiserver".to_owned()),
                    ("endpoint".to_owned(), "https".to_owned()),
                    ("group".to_owned(), "events.k8s.io".to_owned()),
                    ("job".to_owned(), "apiserver".to_owned()),
                    ("namespace".to_owned(), "default".to_owned()),
                    ("resource".to_owned(), "events".to_owned()),
                ]))
            );
        }
        {
            // Checking for existence of `namespace` label should also yield the
            // second time series only.
            let req = test::TestRequest::get().uri("/metrics-samples?namespace").to_request();

            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let result: Vec<Sample> = test::read_body_json(resp).await;
            assert_eq!(result.len(), 1);
            assert_eq!(result[0].metric, "apiserver_request_total");
            assert_eq!(result[0].value, 128.12);
            assert_eq!(result[0].timestamp, 1638873379541);
            assert_eq!(
                result[0].labels,
                HashMap::<String, String>::from_iter(IntoIter::new([
                    ("cluster".to_owned(), "microk8s".to_owned()),
                    ("code".to_owned(), "200".to_owned()),
                    ("component".to_owned(), "apiserver".to_owned()),
                    ("endpoint".to_owned(), "https".to_owned()),
                    ("group".to_owned(), "events.k8s.io".to_owned()),
                    ("job".to_owned(), "apiserver".to_owned()),
                    ("namespace".to_owned(), "default".to_owned()),
                    ("resource".to_owned(), "events".to_owned()),
                ]))
            );
        }
        {
            // and now let's check the previous time series with URL query params
            let req = test::TestRequest::get().uri("/metrics-samples?mock=yes").to_request();

            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let result: Vec<Sample> = test::read_body_json(resp).await;
            assert_eq!(result.len(), 1);
            assert_eq!(result[0].metric, "apiserver_request_total");
            assert_eq!(result[0].value, 123.12);
            assert_eq!(result[0].timestamp, 1638873379541);
            assert_eq!(
                result[0].labels,
                HashMap::<String, String>::from_iter(IntoIter::new([
                    ("mock".to_owned(), "yes".to_owned()),
                    ("group".to_owned(), "events.k8s.io".to_owned()),
                    ("code".to_owned(), "200".to_owned()),
                    ("job".to_owned(), "apiserver".to_owned()),
                    ("cluster".to_owned(), "microk8s".to_owned()),
                    ("component".to_owned(), "apiserver".to_owned()),
                    ("endpoint".to_owned(), "https".to_owned()),
                ]))
            );
        }
        {
            // Now let's just check that we have those 2 time series when no
            // filters are applied.
            let req = test::TestRequest::get().uri("/metrics-samples").to_request();

            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let result: Vec<Sample> = test::read_body_json(resp).await;
            assert_eq!(result.len(), 2);
        }
    }
}
