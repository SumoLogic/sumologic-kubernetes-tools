use std::collections::{HashMap, HashSet};
use std::iter::FromIterator;
use std::net::{IpAddr, Ipv4Addr};
use std::sync::{Mutex, RwLock};

use crate::logs;
use crate::metadata::{get_common_metadata_from_headers, parse_sumo_fields_header_value, Metadata};
use crate::metrics;
use crate::metrics::Sample;
use crate::options;
use crate::time::get_now;
use actix_http::header::HeaderValue;
use actix_web::{http::StatusCode, web, HttpRequest, HttpResponse, Responder};
use rand::distributions::Alphanumeric;
use rand::{thread_rng, Rng};
use serde::{Deserialize, Serialize};

pub mod api;
pub mod otlp;

const DUMMY_ERROR_ID: &str = "E40YU-CU3Q7-RQDM7";

pub struct AppState {
    // Mutexes are necessary to mutate data safely across threads in handlers.
    //
    pub log_stats: RwLock<logs::LogStatsRepository>,
    pub log_messages: RwLock<logs::LogRepository>,

    pub metrics: RwLock<u64>,
    pub metrics_samples: RwLock<HashSet<Sample>>,
    pub metrics_list: RwLock<HashMap<String, u64>>,
    pub metrics_ip_list: RwLock<HashMap<IpAddr, u64>>,
}

impl AppState {
    pub fn new() -> Self {
        return Self {
            log_stats: RwLock::new(logs::LogStatsRepository::new()),
            log_messages: RwLock::new(logs::LogRepository::new()),

            metrics: RwLock::new(0),
            metrics_list: RwLock::new(HashMap::new()),
            metrics_ip_list: RwLock::new(HashMap::new()),
            metrics_samples: RwLock::new(HashSet::new()),
        };
    }
}

impl AppState {
    pub fn add_metrics_result(&self, result: metrics::MetricsHandleResult, opts: &options::Options) {
        {
            let mut metrics = self.metrics.write().unwrap();
            *metrics += result.metrics_count;
        }

        {
            let mut metrics_list = self.metrics_list.write().unwrap();
            for (name, count) in result.metrics_list.iter() {
                *metrics_list.entry(name.clone()).or_insert(0) += count;
            }
        }

        {
            let mut metrics_ip_list = self.metrics_ip_list.write().unwrap();
            for (&ip_address, count) in result.metrics_ip_list.iter() {
                *metrics_ip_list.entry(ip_address).or_insert(0) += count;
            }
        }

        if opts.store_metrics {
            // Replace old data points that represent the same data series
            // (the same metric name and labels) with new ones.
            let mut samples = self.metrics_samples.write().unwrap();
            for s in result.metrics_samples {
                samples.replace(s);
            }
        }
    }
    pub fn add_log_lines<'a>(
        &self,
        lines: impl Iterator<Item = &'a str>,
        metadata: Metadata,
        ipaddr: IpAddr,
        opts: &options::Options,
    ) {
        let mut message_count = 0;
        let mut byte_count = 0;
        let mut log_messages = self.log_messages.write().unwrap();
        for line in lines {
            message_count += 1;
            byte_count += line.len() as u64;
            if opts.store_logs {
                log_messages.add_log_message(line.to_string(), metadata.clone())
            }
        }
        let mut log_stats = self.log_stats.write().unwrap();
        log_stats.update(message_count, byte_count, ipaddr);
    }
}

#[derive(Serialize)]
struct ReceiverErrorErrorsField {
    code: String,
    message: String,
}
#[derive(Serialize)]
struct ReceiverError {
    id: String,
    errors: Vec<ReceiverErrorErrorsField>,
}

pub struct AppMetadata {
    pub url: String,
}

pub struct TerraformState {
    pub fields: Mutex<HashMap<String, String>>,
}

// Reset metrics
pub async fn handler_metrics_reset(app_state: web::Data<AppState>) -> impl Responder {
    *app_state.metrics.write().unwrap() = 0;
    app_state.metrics_list.write().unwrap().clear();
    app_state.metrics_ip_list.write().unwrap().clear();
    app_state.metrics_samples.write().unwrap().clear();

    HttpResponse::Ok().body("All metrics were reset successfully")
}

// List metrics in format: <name>: <count>
pub async fn handler_metrics_list(app_state: web::Data<AppState>) -> impl Responder {
    let mut out = String::new();
    let metrics_list = app_state.metrics_list.read().unwrap();
    for (name, count) in metrics_list.iter() {
        out.push_str(&format!("{}: {}\n", name, count));
    }
    HttpResponse::Ok().body(out)
}

// List metrics in format: <ip_address>: <count>
pub async fn handler_metrics_ips(app_state: web::Data<AppState>) -> impl Responder {
    let mut out = String::new();
    let metrics_ip_list = app_state.metrics_ip_list.read().unwrap();
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

    let samples = &*app_state.metrics_samples.read().unwrap();
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
        app_state.metrics.read().unwrap(),
        app_state.log_stats.read().unwrap().total.message_count,
        app_state.log_stats.read().unwrap().total.byte_count,
    );

    {
        let metrics_ip_list = app_state.metrics_ip_list.read().unwrap();
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
        let log_ipaddr_stats = &app_state.log_stats.read().unwrap().ipaddr;
        if log_ipaddr_stats.len() > 0 {
            let mut logs_ip_count_bytes_string = String::from("# TYPE receiver_mock_logs_bytes_ip_count counter\n");
            let mut logs_ip_count_string = String::from("# TYPE receiver_mock_logs_ip_count counter\n");

            for (ip, val) in log_ipaddr_stats.iter() {
                logs_ip_count_string.push_str(&format!(
                    "receiver_mock_logs_ip_count{{ip_address=\"{}\"}} {}\n",
                    ip, val.message_count
                ));
                logs_ip_count_bytes_string.push_str(&format!(
                    "receiver_mock_logs_bytes_ip_count{{ip_address=\"{}\"}} {}\n",
                    ip, val.byte_count
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

#[derive(Deserialize, Serialize)]
#[serde(rename_all = "camelCase")]
struct TerraformFieldObject {
    field_name: String,
    field_id: String,
    data_type: String,
    state: String,
}

#[derive(Deserialize, Serialize)]
struct TerraformFieldsResponse {
    data: Vec<TerraformFieldObject>,
}

pub async fn handler_terraform_fields(terraform_state: web::Data<TerraformState>) -> impl Responder {
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
        Some(_) => HttpResponse::with_body(
            StatusCode::OK,
            json_str!({
                id: "E40YU-CU3Q7-RQDMO",
                errors: [
                    {
                        code: "field:already_exists",
                        message: "Field with the given name already exists"
                    }
                ]
            }),
        )
        .map_into_boxed_body(),

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
                Some(_) => HttpResponse::with_body(
                    StatusCode::BAD_REQUEST,
                    json_str!({
                        id: "E40YU-CU3Q7-RQDMO",
                        errors: [
                            {
                                code: "field:already_exists",
                                message: "Field with the given name already exists"
                            }
                        ]

                    }),
                )
                .map_into_boxed_body(),
            }
        }
    }
}

pub async fn handler_receiver(
    req: HttpRequest,
    body: web::Bytes,
    app_state: web::Data<AppState>,
    opts: web::Data<options::Options>,
) -> impl Responder {
    let remote_address = get_address(&req);
    // actix automatically decompresses body for us.
    let string_body = String::from_utf8(body.to_vec()).unwrap();
    let lines = string_body.trim().lines();

    let content_type = get_content_type(&req);

    // parse the value of the X-Sumo-* headers, excluding X-Sumo-Fields, which is handled separately
    // TODO: use the metadata for metrics
    let metadata = match get_common_metadata_from_headers(req.headers()) {
        Ok(metadata) => metadata,
        Err(error) => return HttpResponse::BadRequest().body(error.to_string()),
    };

    if let Some(response) = try_dropping_data(&opts, &content_type) {
        return response;
    }

    match content_type.as_str() {
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
            // parse X-Sumo-Fields for metadata
            let mut metadata = metadata;
            match req.headers().get("x-sumo-fields") {
                Some(header_value) => match header_value.to_str() {
                    Ok(header_value_str) => match parse_sumo_fields_header_value(header_value_str) {
                        Ok(fields_metadata) => metadata.extend(fields_metadata),
                        Err(_) => return HttpResponse::BadRequest().body("Unable to parse X-Sumo-Fields header value"),
                    },
                    Err(_) => return HttpResponse::BadRequest().body("Unable to parse X-Sumo-Fields header value"),
                },
                None => (),
            };
            app_state.add_log_lines(lines.clone(), metadata, remote_address, &opts);
            if opts.print.logs {
                for line in lines.clone() {
                    println!("log => {}", line);
                }
            }
        }

        &_ => {
            return get_invalid_header_response(&content_type);
        }
    }

    HttpResponse::Ok().body("")
}

fn try_dropping_data(opts: &web::Data<options::Options>, content_type: &str) -> Option<HttpResponse> {
    let mut rng = rand::thread_rng();
    let number: i64 = rng.gen_range(0..100);
    if number < opts.drop_rate {
        let msg = format!("Dropping data for {}", content_type);
        println!("{}", msg);
        return Some(HttpResponse::InternalServerError().body(msg));
    }

    None
}

fn get_address(req: &HttpRequest) -> IpAddr {
    // Don't fail when we can't read remote address.
    // Default to localhost and just ingest what was sent.
    let localhost: std::net::SocketAddr = std::net::SocketAddr::new(IpAddr::V4(Ipv4Addr::new(127, 0, 0, 1)), 0);
    req.peer_addr().unwrap_or(localhost).ip()
}

fn get_content_type(req: &HttpRequest) -> String {
    let empty_header = HeaderValue::from_str("").unwrap();
    req.headers()
        .get("content-type")
        .unwrap_or(&empty_header)
        .to_str()
        .unwrap()
        .to_string()
}

fn get_invalid_header_response(content_type: &str) -> HttpResponse {
    HttpResponse::build(StatusCode::BAD_REQUEST).json(ReceiverError {
        id: String::from(DUMMY_ERROR_ID),
        errors: vec![ReceiverErrorErrorsField {
            code: String::from("header:invalid"),
            message: format!("Invalid Content-Type header: {}", content_type),
        }],
    })
}

// Data structures and handlers for logs endpoints start here
#[derive(Deserialize)]
pub struct LogsParams {
    #[serde(default = "default_from_ts")]
    from_ts: u64,
    #[serde(default = "default_to_ts")]
    to_ts: u64,
}

// Unfortunately serde doesn't allow defaults which are simple constants
fn default_from_ts() -> u64 {
    return 0;
}

fn default_to_ts() -> u64 {
    return u64::MAX;
}

#[derive(Serialize, Deserialize)]
pub struct LogsCountResponse {
    count: usize,
}

// Returns the number of logs received in a given timestamp range
pub async fn handler_logs_count(
    app_state: web::Data<AppState>,
    web::Query(params): web::Query<LogsParams>,
    web::Query(all_params): web::Query<HashMap<String, String>>,
    opts: web::Data<options::Options>,
) -> impl Responder {
    if !opts.store_logs {
        return HttpResponse::NotImplemented().body("Use the --store-logs flag to enable this endpoint");
    }
    // all_params has all the parameters, so we need to remove the fixed ones
    let fixed_params: HashSet<&str> = HashSet::from_iter(vec!["from_ts", "to_ts"].into_iter());
    let metadata_params: HashMap<&str, &str> = all_params
        .iter()
        .filter(|(key, _)| !fixed_params.contains(key.as_str()))
        .map(|(key, value)| (key.as_str(), value.as_str()))
        .collect();
    let count = app_state
        .log_messages
        .read()
        .unwrap()
        .get_message_count(params.from_ts, params.to_ts, metadata_params);

    HttpResponse::Ok().json(LogsCountResponse { count })
}

pub async fn handler_dump(body: web::Bytes) -> impl Responder {
    let string_body = String::from_utf8(body.to_vec()).unwrap();
    println!("dump: {}", string_body);
    HttpResponse::Ok().body("")
}

pub fn print_request_headers(
    method: &http::Method,
    version: http::Version,
    uri: &http::Uri,
    headers: &actix_http::header::HeaderMap,
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
        let metrics = app_state.metrics.read().unwrap();
        let log_stats = app_state.log_stats.read().unwrap();

        // TODO: make this print metrics per minute (as DPM) and logs
        // per second, regardless of used interval
        // ref: https://github.com/SumoLogic/sumologic-kubernetes-tools/issues/57
        println!(
            "{} Metrics: {:10.} Logs: {:10.}; {:6.6} MB/s",
            now,
            *metrics - p_metrics,
            log_stats.total.message_count - p_logs,
            ((log_stats.total.byte_count - p_logs_bytes) as f64) / ((now - ts) as f64) / (1e6 as f64)
        );

        ts = now;
        p_metrics = *metrics;
        p_logs = log_stats.total.message_count;
        p_logs_bytes = log_stats.total.byte_count;
    })
}

#[cfg(test)]
mod tests_terraform {
    use super::*;
    use actix_rt;
    use actix_web::{test, web, App};

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
            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let body = test::read_body(resp).await;
            assert_eq!(
                body,
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
            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let body = test::read_body(resp).await;
            assert_eq!(
                body,
                json_str!({
                    source: {
                      url: "http://hostname:3000/terraform"
                    }
                })
            );
        }
        {
            let req = test::TestRequest::get().uri("/different_route").to_request();
            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 404);

            let body = test::read_body(resp).await;
            assert_eq!(body, "");
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
            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let body = test::read_body(resp).await;
            assert_eq!(body, r#"{"quota":200,"remaining":100}"#);
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
            let req = test::TestRequest::get()
                .uri("/terraform/api/v1/fields")
                .to_request();
            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let body = test::read_body(resp).await;
            assert_eq!(
                body,
                json_str!({
                    data: []
                }),
            );
        }

        {
            let req = test::TestRequest::get()
                .uri("/terraform/api/v1/fields/dummyID123")
                .to_request();
            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 404);

            let body = test::read_body(resp).await;
            assert_eq!(
                body,
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
            let req = test::TestRequest::get()
                .uri("/terraform/api/v1/fields")
                .to_request();
            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let body = test::read_body(resp).await;
            assert_eq!(
                body,
                json_str!({
                    data: []
                })
            );
        }

        {
            let field_id: String;
            {
                // Create a field ...
                let req = test::TestRequest::post()
                    .uri("/terraform/api/v1/fields")
                    .set_json(&TerraformFieldCreateRequest {
                        field_name: String::from("dummyID123"),
                    })
                    .to_request();
                let resp = test::call_service(&mut app, req).await;
                assert_eq!(resp.status(), 200);

                let body: TerraformFieldObject = test::read_body_json(resp).await;
                assert_eq!(body.field_name, "dummyID123");
                assert_eq!(body.data_type, "String");
                assert_eq!(body.state, "Enabled");
                assert_ne!(body.field_id, "");
                field_id = body.field_id;
            }

            // ... and check it exists
            {
                let req = test::TestRequest::get()
                    .uri("/terraform/api/v1/fields")
                    .to_request();
                let resp = test::call_service(&mut app, req).await;
                assert_eq!(resp.status(), 200);

                let body: TerraformFieldsResponse = test::read_body_json(resp).await;
                assert_eq!(body.data.len(), 1);
                assert_eq!(body.data[0].field_name, "dummyID123");
                assert_eq!(body.data[0].data_type, "String");
                assert_eq!(body.data[0].state, "Enabled");
                assert_eq!(body.data[0].field_id, field_id);
            }
        }
    }
}

#[cfg(test)]
mod tests_metrics {
    use super::*;
    use actix_rt;
    use actix_web::{test, web, App};
    use std::iter::FromIterator;

    #[actix_rt::test]
    async fn default_handler_protobuf_unsupported_invalid_header() {
        let mut app = test::init_service(App::new().default_service(web::get().to(handler_receiver))).await;

        {
            let req = test::TestRequest::post()
                .uri("/")
                .insert_header(("Content-Type", "application/x-protobuf"))
                .to_request();
            let resp = test::call_service(&mut app, req).await;

            assert_eq!(resp.status(), 500);
        }
    }

    #[actix_rt::test]
    async fn test_handler_metrics_reset() {
        let mut metrics_list = HashMap::new();
        metrics_list.insert(String::from("mem_active"), 1000);
        metrics_list.insert(String::from("mem_free"), 2000);

        let app_state = AppState::new();
        *app_state.metrics.write().unwrap() = 3000;
        *app_state.metrics_list.write().unwrap() = metrics_list;
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
            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let body = test::read_body(resp).await;
            assert_eq!(
                web::Bytes::from_static(
                    b"# TYPE receiver_mock_metrics_count counter\n\
                 receiver_mock_metrics_count 3000\n\
                 # TYPE receiver_mock_logs_count counter\n\
                 receiver_mock_logs_count 0\n\
                 # TYPE receiver_mock_logs_bytes_count counter\n\
                 receiver_mock_logs_bytes_count 0\n",
                ),
                body,
            );
        }
        {
            let req = test::TestRequest::post().uri("/metrics-reset").to_request();
            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let body = test::read_body(resp).await;
            assert_eq!(body, "All metrics were reset successfully");
        }
        {
            let req = test::TestRequest::get().uri("/metrics").to_request();
            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let body = test::read_body(resp).await;
            assert_eq!(
                web::Bytes::from_static(
                    b"# TYPE receiver_mock_metrics_count counter\n\
                  receiver_mock_metrics_count 0\n\
                  # TYPE receiver_mock_logs_count counter\n\
                  receiver_mock_logs_count 0\n\
                  # TYPE receiver_mock_logs_bytes_count counter\n\
                  receiver_mock_logs_bytes_count 0\n",
                ),
                body,
            );
        }
    }

    #[actix_rt::test]
    async fn test_handler_metrics_list() {
        let mut metrics_list = HashMap::new();
        metrics_list.insert(String::from("mem_free"), 2000);

        let app_state = AppState::new();
        *app_state.metrics.write().unwrap() = 2000;
        *app_state.metrics_list.write().unwrap() = metrics_list;
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
            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let body = test::read_body(resp).await;

            // Testing multiple metric names being returned properly would be nice
            // but since we're using a hashmap we'd need to sort the lines from the byte
            // buffer that we receive and ain't that easy (but definitely doable).
            assert_eq!(body, "mem_free: 2000\n");
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
            delay_time: std::time::Duration::from_secs(0),
            drop_rate: 0,
            store_metrics: true,
            store_logs: true,
        };

        let mut app = test::init_service(
            actix_web::App::new()
                .app_data(web::Data::new(opts.clone()))
                .app_data(web_data_app_state.clone()) // Mutable shared state
                .route("/metrics-samples", web::get().to(handler_metrics_samples))
                .default_service(web::get().to(handler_receiver)),
        )
        .await;

        {
            let req = test::TestRequest::post().uri("/")
            .set_payload(r#"apiserver_request_total{cluster="microk8s",mock="yes",code="200",component="apiserver",endpoint="https",group="events.k8s.io",job="apiserver"} 123.12 1638873379541"#)
            .insert_header(("Content-Type", "application/vnd.sumologic.prometheus"))
            .to_request();

            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let body = test::read_body(resp).await;
            assert_eq!(body, web::Bytes::from_static(b""));
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
                HashMap::<String, String>::from_iter(
                    vec![
                        ("mock".to_owned(), "yes".to_owned()),
                        ("group".to_owned(), "events.k8s.io".to_owned()),
                        ("code".to_owned(), "200".to_owned()),
                        ("job".to_owned(), "apiserver".to_owned()),
                        ("cluster".to_owned(), "microk8s".to_owned()),
                        ("component".to_owned(), "apiserver".to_owned()),
                        ("endpoint".to_owned(), "https".to_owned()),
                    ]
                    .into_iter()
                )
            );
        }
        {
            // Another request with a different time series (different labels set)
            // should produce a different/new time series
            let req = test::TestRequest::post().uri("/")
            .set_payload(r#"apiserver_request_total{cluster="microk8s",code="200",component="apiserver",endpoint="https",group="events.k8s.io",job="apiserver",namespace="default",resource="events"} 128.12 1638873379541"#)
            .insert_header(("Content-Type", "application/vnd.sumologic.prometheus"))
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
                HashMap::<String, String>::from_iter(
                    vec![
                        ("cluster".to_owned(), "microk8s".to_owned()),
                        ("code".to_owned(), "200".to_owned()),
                        ("component".to_owned(), "apiserver".to_owned()),
                        ("endpoint".to_owned(), "https".to_owned()),
                        ("group".to_owned(), "events.k8s.io".to_owned()),
                        ("job".to_owned(), "apiserver".to_owned()),
                        ("namespace".to_owned(), "default".to_owned()),
                        ("resource".to_owned(), "events".to_owned()),
                    ]
                    .into_iter()
                )
            );
        }
        {
            // Checking for existence of `namespace` label should also yield the
            // second time series only.
            let req = test::TestRequest::get()
                .uri("/metrics-samples?namespace")
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
                HashMap::<String, String>::from_iter(
                    vec![
                        ("cluster".to_owned(), "microk8s".to_owned()),
                        ("code".to_owned(), "200".to_owned()),
                        ("component".to_owned(), "apiserver".to_owned()),
                        ("endpoint".to_owned(), "https".to_owned()),
                        ("group".to_owned(), "events.k8s.io".to_owned()),
                        ("job".to_owned(), "apiserver".to_owned()),
                        ("namespace".to_owned(), "default".to_owned()),
                        ("resource".to_owned(), "events".to_owned()),
                    ]
                    .into_iter()
                )
            );
        }
        {
            // and now let's check the previous time series with URL query params
            let req = test::TestRequest::get()
                .uri("/metrics-samples?mock=yes")
                .to_request();

            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let result: Vec<Sample> = test::read_body_json(resp).await;
            assert_eq!(result.len(), 1);
            assert_eq!(result[0].metric, "apiserver_request_total");
            assert_eq!(result[0].value, 123.12);
            assert_eq!(result[0].timestamp, 1638873379541);
            assert_eq!(
                result[0].labels,
                HashMap::<String, String>::from_iter(
                    vec![
                        ("mock".to_owned(), "yes".to_owned()),
                        ("group".to_owned(), "events.k8s.io".to_owned()),
                        ("code".to_owned(), "200".to_owned()),
                        ("job".to_owned(), "apiserver".to_owned()),
                        ("cluster".to_owned(), "microk8s".to_owned()),
                        ("component".to_owned(), "apiserver".to_owned()),
                        ("endpoint".to_owned(), "https".to_owned()),
                    ]
                    .into_iter()
                )
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

#[cfg(test)]
mod tests_logs {
    use super::*;
    use actix_rt;
    use actix_web::{test, web, App};

    #[actix_rt::test]
    async fn default_handler_protobuf_unsupported_invalid_header() {
        let mut app = test::init_service(App::new().default_service(web::get().to(handler_receiver))).await;

        {
            let req = test::TestRequest::post()
                .uri("/")
                .insert_header(("Content-Type", "application/x-protobuf"))
                .to_request();
            let resp = test::call_service(&mut app, req).await;

            assert_eq!(resp.status(), 500);
        }
    }

    #[actix_rt::test]
    async fn test_handler_logs_count() {
        let x_sumo_fields_values = [
            "namespace=default, deployment=collection-kube-state-metrics, node=sumologic-control-plane",
            "namespace=sumologic, deployment=collection-kube-state-metrics, node=sumologic-control-plane",
            "namespace=kube-system, statefulset=collection-fluentd-metrics",
        ];
        let timestamps = [1, 5, 8];
        let raw_logs: Vec<_> = timestamps
            .iter()
            .map(|ts| format!("{{\"log\": \"Log message\", \"timestamp\": {}}}", ts))
            .collect();
        let app_state = AppState::new();
        let app_data = web::Data::new(app_state);
        let opts = options::Options {
            print: options::Print {
                logs: false,
                headers: false,
                metrics: false,
            },
            delay_time: std::time::Duration::from_secs(0),
            drop_rate: 0,
            store_metrics: true,
            store_logs: true,
        };

        let mut app = test::init_service(
            App::new()
                .app_data(web::Data::new(opts.clone()))
                .app_data(app_data.clone()) // Mutable shared state
                .route("/logs/count", web::get().to(handler_logs_count))
                .default_service(web::get().to(handler_receiver)),
        )
        .await;

        // invalid x-sumo-fields results in a 400
        {
            let log_payload = raw_logs[0].clone();
            let x_sumo_fields_value = ",no_equals_sign";
            let req = test::TestRequest::post()
                .uri("/")
                .set_payload(log_payload)
                .insert_header(("Content-Type", "application/x-www-form-urlencoded"))
                .insert_header(("X-Sumo-Fields", x_sumo_fields_value))
                .to_request();

            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 400);

            let body = test::read_body(resp).await;
            assert_eq!(
                body,
                web::Bytes::from_static(b"Unable to parse X-Sumo-Fields header value")
            );
        }

        // add logs with metadata
        {
            for i in 0..raw_logs.len() {
                let log_payload = raw_logs[i].clone();
                let x_sumo_fields_value = x_sumo_fields_values[i];
                let req = test::TestRequest::post()
                    .uri("/")
                    .set_payload(log_payload)
                    .insert_header(("Content-Type", "application/x-www-form-urlencoded"))
                    .insert_header(("X-Sumo-Fields", x_sumo_fields_value))
                    .insert_header(("X-Sumo-Host", "localhost"))
                    .insert_header(("X-Sumo-Category", "category"))
                    .insert_header(("X-Sumo-Name", "name"))
                    .to_request();

                let resp = test::call_service(&mut app, req).await;
                assert_eq!(resp.status(), 200);
            }
        }

        // count all the logs
        {
            let req = test::TestRequest::get().uri("/logs/count").to_request();
            let resp = test::call_service(&mut app, req).await;

            let response_body: LogsCountResponse = test::read_body_json(resp).await;

            assert_eq!(response_body.count, 3);
        }

        // from_ts is inclusive
        {
            let req = test::TestRequest::get()
                .uri("/logs/count?from_ts=5")
                .to_request();
            let resp = test::call_service(&mut app, req).await;

            let response_body: LogsCountResponse = test::read_body_json(resp).await;

            assert_eq!(response_body.count, 2);
        }

        // to_ts is exclusive
        {
            let req = test::TestRequest::get().uri("/logs/count?to_ts=5").to_request();
            let resp = test::call_service(&mut app, req).await;

            let response_body: LogsCountResponse = test::read_body_json(resp).await;

            assert_eq!(response_body.count, 1);
        }

        // normal metadata query
        {
            let req = test::TestRequest::get()
                .uri("/logs/count?deployment=collection-kube-state-metrics")
                .to_request();
            let resp = test::call_service(&mut app, req).await;

            let response_body: LogsCountResponse = test::read_body_json(resp).await;

            assert_eq!(response_body.count, 2);
        }

        // X-Sumo-* fields
        {
            let req = test::TestRequest::get()
                .uri("/logs/count?_sourceName=name&_sourceHost=localhost&_sourceCategory=category")
                .to_request();
            let resp = test::call_service(&mut app, req).await;

            let response_body: LogsCountResponse = test::read_body_json(resp).await;

            assert_eq!(response_body.count, 3);
        }

        // wildcard query
        {
            let req = test::TestRequest::get()
                .uri("/logs/count?namespace=")
                .to_request();
            let resp = test::call_service(&mut app, req).await;

            let response_body: LogsCountResponse = test::read_body_json(resp).await;

            assert_eq!(response_body.count, 3);
        }

        // everything at once
        {
            let req = test::TestRequest::get()
                .uri("/logs/count?namespace=&deployment=collection-kube-state-metrics&from_ts=5&to_ts=10")
                .to_request();
            let resp = test::call_service(&mut app, req).await;

            let response_body: LogsCountResponse = test::read_body_json(resp).await;

            assert_eq!(response_body.count, 1);
        }
    }
}
