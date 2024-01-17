use std::collections::HashMap;

use actix_web::{web, HttpResponse, Responder};

use crate::{options, traces};

use super::AppState;

pub async fn handler_get_spans(
    app_state: web::Data<AppState>,
    web::Query(params): web::Query<HashMap<String, String>>,
    opts: web::Data<options::Options>,
) -> impl Responder {
    if !opts.store_traces {
        return HttpResponse::NotImplemented().body("");
    }

    let spans = &*app_state.spans_list.read().unwrap();
    let response = traces::filter_spans(spans.values(), params);

    HttpResponse::Ok().json(response)
}

pub async fn handler_get_traces(
    app_state: web::Data<AppState>,
    web::Query(params): web::Query<HashMap<String, String>>,
    opts: web::Data<options::Options>,
) -> impl Responder {
    if !opts.store_traces {
        return HttpResponse::NotImplemented().body("");
    }

    let spans = &*app_state.spans_list.read().unwrap();
    let traces = &*app_state.traces_list.read().unwrap();
    let response = traces::filter_traces(traces.values(), &spans, params);

    HttpResponse::Ok().json(response)
}
