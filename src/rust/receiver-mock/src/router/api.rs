pub mod v1 {
    use actix_http::body::Body;
    use actix_web::{HttpRequest, Responder};
    use base64;
    use serde::Serialize;

    #[derive(Serialize)]
    struct CollectorRegisterRespone {
        collector_credential_id: String,
        collector_credential_key: String,
        collector_id: String,
        collector_name: String,
    }

    pub async fn handler_collector_register(req: HttpRequest) -> impl Responder {
        let header_value = match req.headers().get("Authorization") {
            Some(v) => v,
            None => return actix_http::Response::BadRequest().body(Body::Empty),
        };

        let val_str = match header_value.to_str() {
            Ok(v) => v,
            Err(_) => return actix_http::Response::BadRequest().body(Body::Empty),
        };

        let val = match val_str.strip_prefix("Basic ") {
            Some(v) => v,
            None => return actix_http::Response::BadRequest().body(Body::Empty),
        };

        // For now the token is only checked if it can be decoded successfully.
        let _decoded = match base64::decode(val) {
            Ok(v) => v,
            Err(_) => return actix_http::Response::BadRequest().body(Body::Empty),
        };

        actix_http::Response::Ok().json(CollectorRegisterRespone {
            collector_credential_id: String::from("eeeQShpym1Szkza33333"),
            collector_credential_key: String::from("eeef3dD3nBUorbP6s3NFTya0JwLZ0FosrIsRREumZoWXEt7szGoJViwbdc5lfHq73Slsv7OctRzlvTfMLyexLULI8mYe8gFhmUZS75BhgcvqFZEfWb2Z6OsFnOxmAAAA"),
            collector_id: String::from("000000000111AAA3"),
            collector_name: String::from("collector-test-123456123123"),
        })
    }

    pub async fn handler_collector_heartbeat() -> impl Responder {
        actix_http::Response::NoContent().body(Body::Empty)
    }
}
