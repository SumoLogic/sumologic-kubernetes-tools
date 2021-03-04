use std::collections::HashMap;
use std::net::{IpAddr, Ipv4Addr};
use std::sync::Mutex;

use actix_http::http;
use actix_web::{http::StatusCode, web, HttpRequest, HttpResponse, Responder};
use bytes;
use chrono::Duration;
use rand::distributions::Alphanumeric;
use rand::{thread_rng, Rng};
use serde_derive::{Deserialize, Serialize};

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

pub async fn handler_terraform_fields(
    terraform_state: web::Data<TerraformState>,
) -> impl Responder {

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

    web::Json(TerraformFieldsResponse {
        data: res.collect(),
    })
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
    let exists = fields.iter().find_map(|(id, name)| {
        if name == &requested_name {
            Some(id)
        } else {
            None
        }
    });

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
    let localhost: std::net::SocketAddr =
        std::net::SocketAddr::new(IpAddr::V4(Ipv4Addr::new(127, 0, 0, 1)), 0);
    let remote_address = req.peer_addr().unwrap_or(localhost).ip();

    let body_length = body.len() as u64;
    // actix automatically decompresses body for us.
    let string_body = String::from_utf8(body.to_vec()).unwrap();
    let lines = string_body.trim().lines();

    let headers = req.headers();
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

#[cfg(test)]
mod tests {
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
            let req = test::TestRequest::get()
                .uri("/different_route")
                .to_request();
            let mut resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 404);

            let bytes = test::load_stream(resp.take_body().into_stream()).await;
            assert_eq!(bytes.unwrap(), "");
        }
    }

    #[actix_rt::test]
    async fn test_handler_metrics_reset() {
        let mut metrics_list = HashMap::new();
        metrics_list.insert(String::from("mem_active"), 1000);
        metrics_list.insert(String::from("mem_free"), 2000);
        let app_state = web::Data::new(AppState {
            metrics: Mutex::new(3000),
            logs: Mutex::new(0),
            logs_bytes: Mutex::new(0),
            metrics_list: Mutex::new(metrics_list),
            metrics_ip_list: Mutex::new(HashMap::new()),
            logs_ip_list: Mutex::new(HashMap::new()),
        });

        let mut app = test::init_service(
            App::new()
                .app_data(app_state.clone()) // Mutable shared state
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
        let app_state = web::Data::new(AppState {
            metrics: Mutex::new(2000),
            logs: Mutex::new(0),
            logs_bytes: Mutex::new(0),
            metrics_list: Mutex::new(metrics_list),
            metrics_ip_list: Mutex::new(HashMap::new()),
            logs_ip_list: Mutex::new(HashMap::new()),
        });

        let mut app = test::init_service(
            App::new()
                .app_data(app_state.clone()) // Mutable shared state
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
    async fn test_handler_terraform_fields_quota() {
        let app_metadata = web::Data::new(AppMetadata {
            url: String::from("http://hostname:3000/receiver"),
        });

        let app_state = web::Data::new(AppState {
            metrics: Mutex::new(0),
            logs: Mutex::new(0),
            logs_bytes: Mutex::new(0),
            metrics_list: Mutex::new(HashMap::new()),
            metrics_ip_list: Mutex::new(HashMap::new()),
            logs_ip_list: Mutex::new(HashMap::new()),
        });

        let mut app = test::init_service(
            App::new()
                .app_data(app_state.clone()) // Mutable shared state
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

        let app_state = web::Data::new(AppState {
            metrics: Mutex::new(0),
            logs: Mutex::new(0),
            logs_bytes: Mutex::new(0),
            metrics_list: Mutex::new(HashMap::new()),
            metrics_ip_list: Mutex::new(HashMap::new()),
            logs_ip_list: Mutex::new(HashMap::new()),
        });

        let terraform_state = web::Data::new(TerraformState {
            fields: Mutex::new(HashMap::new()),
        });

        let mut app = test::init_service(
            App::new()
                .app_data(app_state.clone()) // Mutable shared state
                .service(
                    web::scope("/terraform")
                        .app_data(app_metadata.clone())
                        .app_data(terraform_state.clone())
                        .route(
                            "/api/v1/fields/{field}",
                            web::get().to(handler_terraform_field),
                        )
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
            let req = test::TestRequest::get()
                .uri("/terraform/api/v1/fields")
                .to_request();
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
            let req = test::TestRequest::get()
                .uri("/terraform/api/v1/fields")
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
    }
}
