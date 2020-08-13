use std::net::IpAddr;
use std::sync::{Arc, Mutex};

use crate::statistics::Statistics;

pub fn handle_carbon2(
    lines: std::str::Lines,
    address: IpAddr,
    stats: &Arc<Mutex<Statistics>>,
) {
    let mut stats = stats.lock().unwrap();

    for line in lines {
        let mut split = line.split("  ");
        let intrinsic_metrics = split.nth(0).unwrap();
        for metric in intrinsic_metrics.split(" ") {
            let metric_name = metric.split("=").nth(0).unwrap().to_string();
            if metric_name == "metric" {
                let saved_metric = (*stats).metrics_list.entry(metric_name).or_insert(0);
                *saved_metric += 1;
                (*stats).metrics += 1;

                break;
            }
        }

        let metrics_ip_list = (*stats).metrics_ip_list.entry(address).or_insert(0);
        *metrics_ip_list += 1;
    }
}

pub fn handle_prometheus(
    lines: std::str::Lines,
    address: IpAddr,
    stats: &Arc<Mutex<Statistics>>,
) {
    let mut stats = stats.lock().unwrap();

    for line in lines {
        let metric_name = line.split("{").nth(0).unwrap().to_string();
        let saved_metric = (*stats).metrics_list.entry(metric_name).or_insert(0);

        *saved_metric += 1;
        (*stats).metrics += 1;

        let metrics_ip_list = (*stats).metrics_ip_list.entry(address).or_insert(0);
        *metrics_ip_list += 1;
    }
}
