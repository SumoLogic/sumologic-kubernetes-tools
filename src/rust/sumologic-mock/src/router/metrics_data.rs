use std::collections::HashMap;

use crate::metrics;
use crate::options;
use actix_web::{web, HttpResponse, Responder};

use super::AppState;

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
    let response = metrics::sample::filter_samples(samples, params);

    HttpResponse::Ok().json(response)
}
