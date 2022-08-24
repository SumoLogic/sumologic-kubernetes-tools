use std::collections::HashMap;
use std::sync::Mutex;

use actix_web::{http::StatusCode, web, HttpResponse, Responder};
use rand::distributions::Alphanumeric;
use rand::{thread_rng, Rng};
use serde::{Deserialize, Serialize};

use super::AppMetadata;

pub struct TerraformState {
    pub fields: Mutex<HashMap<String, String>>,
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

#[cfg(test)]
mod test {
    use crate::router::AppState;

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
