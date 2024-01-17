pub mod v1 {
    use actix_web::{HttpRequest, HttpResponse, Responder};
    use base64::{engine::general_purpose as b64, Engine as _};
    use serde::{Deserialize, Serialize};

    #[derive(Deserialize, Serialize)]
    pub(crate) struct CollectorRegisterRespone {
        #[serde(rename = "collectorCredentialId")]
        collector_credential_id: String,
        #[serde(rename = "collectorCredentialKey")]
        collector_credential_key: String,
        #[serde(rename = "collectorId")]
        collector_id: String,
        #[serde(rename = "collectorName")]
        collector_name: String,
    }

    pub async fn handler_collector_register(req: HttpRequest) -> impl Responder {
        let header_value = match req.headers().get("Authorization") {
            Some(v) => v,
            None => return HttpResponse::BadRequest().finish(),
        };

        let val_str = match header_value.to_str() {
            Ok(v) => v,
            Err(_) => return HttpResponse::BadRequest().finish(),
        };

        if val_str.starts_with("Basic ") {
            let val = match val_str.strip_prefix("Basic ") {
                Some(v) => v,
                None => return HttpResponse::Unauthorized().finish(),
            };

            match b64::STANDARD.decode(val) {
                Ok(v) => v,
                Err(_) => {
                    return HttpResponse::Unauthorized().finish();
                }
            };
        } else if val_str.starts_with("Bearer ") {
            match val_str.strip_prefix("Bearer ") {
                Some(v) => v,
                None => return HttpResponse::Unauthorized().finish(),
            };
        } else {
            return HttpResponse::Unauthorized().finish();
        }

        HttpResponse::Ok().json(CollectorRegisterRespone {
            collector_credential_id: String::from("eeeQShpym1Szkza33333"),
            collector_credential_key: String::from("eeef3dD3nBUorbP6s3NFTya0JwLZ0FosrIsRREumZoWXEt7szGoJViwbdc5lfHq73Slsv7OctRzlvTfMLyexLULI8mYe8gFhmUZS75BhgcvqFZEfWb2Z6OsFnOxmAAAA"),
            collector_id: String::from("000000000111AAA3"),
            collector_name: String::from("collector-test-123456123123"),
        })
    }

    pub async fn handler_collector_heartbeat() -> impl Responder {
        HttpResponse::NoContent().finish()
    }

    pub async fn handler_collector_metadata() -> impl Responder {
        HttpResponse::Ok().finish()
    }
}

#[cfg(test)]
mod tests_api {
    use crate::options;
    use crate::router;
    use actix_rt;
    use actix_web::{test, web, App};

    #[actix_rt::test]
    async fn test_api_v1_collector_register() {
        let opts = options::Options {
            print: options::Print {
                logs: false,
                headers: false,
                metrics: false,
                spans: false,
            },
            delay_time: std::time::Duration::from_secs(0),
            drop_rate: 0,
            store_traces: false,
            store_metrics: false,
            store_logs: false,
        };

        let mut app = test::init_service(App::new().app_data(web::Data::new(opts.clone())).service(
            web::scope("/api/v1").route(
                "/collector/register",
                web::post().to(router::api::v1::handler_collector_register),
            ),
        ))
        .await;

        {
            // No Authorization header returns a 400
            let req = test::TestRequest::post()
                .uri("/api/v1/collector/register")
                .to_request();

            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 400);

            let body = test::read_body(resp).await;
            assert_eq!(body, "");
        }
        {
            // Invalid token in Authorization header returns a 401
            let req = test::TestRequest::post()
                .uri("/api/v1/collector/register")
                .insert_header(("Authorization", "Basic xyz"))
                .to_request();

            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 401);

            let body = test::read_body(resp).await;
            assert_eq!(body, "");
        }
        {
            // Decodable token returns a 200 with JSON payload
            let req = test::TestRequest::post()
                .uri("/api/v1/collector/register")
                .insert_header(("Authorization", "Basic ZHVtbXk6bXlwYXNzd29yZA=="))
                .to_request();

            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let _result: router::api::v1::CollectorRegisterRespone = test::read_body_json(resp).await;
        }
        {
            // Bearer token returns 200
            let req = test::TestRequest::post()
                .uri("/api/v1/collector/register")
                .insert_header(("Authorization", "Bearer xyz"))
                .to_request();

            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let _result: router::api::v1::CollectorRegisterRespone = test::read_body_json(resp).await;
        }
    }

    #[actix_rt::test]
    async fn test_api_v1_collector_heartbeat() {
        let app_data = web::Data::new(router::AppState::new());
        let opts = options::Options {
            print: options::Print {
                logs: false,
                headers: false,
                metrics: false,
                spans: false,
            },
            delay_time: std::time::Duration::from_secs(0),
            drop_rate: 0,
            store_traces: false,
            store_metrics: false,
            store_logs: false,
        };

        let mut app = test::init_service(
            App::new()
                .app_data(web::Data::new(opts.clone()))
                .app_data(app_data.clone()) // Mutable shared state
                .service(web::scope("/api/v1").route(
                    "/collector/heartbeat",
                    web::post().to(router::api::v1::handler_collector_heartbeat),
                )),
        )
        .await;

        {
            let req = test::TestRequest::post()
                .uri("/api/v1/collector/heartbeat")
                .to_request();

            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 204);

            let body = test::read_body(resp).await;
            assert_eq!(body, "");
        }
    }

    #[actix_rt::test]
    async fn test_api_v1_collector_metadata() {
        let app_data = web::Data::new(router::AppState::new());
        let opts = options::Options {
            print: options::Print {
                logs: false,
                headers: false,
                metrics: false,
                spans: false,
            },
            delay_time: std::time::Duration::from_secs(0),
            drop_rate: 0,
            store_traces: false,
            store_metrics: false,
            store_logs: false,
        };

        let mut app = test::init_service(
            App::new()
                .app_data(web::Data::new(opts.clone()))
                .app_data(app_data.clone()) // Mutable shared state
                .service(web::scope("/api/v1").route(
                    "/collector/metadata",
                    web::post().to(router::api::v1::handler_collector_metadata),
                )),
        )
        .await;

        {
            let req = test::TestRequest::post()
                .uri("/api/v1/collector/metadata")
                .to_request();

            let resp = test::call_service(&mut app, req).await;
            assert_eq!(resp.status(), 200);

            let body = test::read_body(resp).await;
            assert_eq!(body, "");
        }
    }
}
