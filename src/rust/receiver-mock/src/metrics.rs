use std::net::IpAddr;
use std::sync::{Arc, Mutex};

use crate::print;
use crate::statistics::Statistics;

// Handle metrics in Carbon2.0 format
// Reference: https://help.sumologic.com/Metrics/Introduction-to-Metrics/Metric-Formats#carbon-2-0
pub fn handle_carbon2(
    lines: std::str::Lines,
    address: IpAddr,
    stats: &Arc<Mutex<Statistics>>,
    print_opts: print::Options,
) {
    let mut stats = stats.lock().unwrap();

    for line in lines {
        if print_opts.print_metrics {
            println!("metric => {}", line);
        }
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

// Handle metrics in Graphite format
// Reference: https://help.sumologic.com/Metrics/Introduction-to-Metrics/Metric-Formats#graphite
pub fn handle_graphite(
    lines: std::str::Lines,
    address: IpAddr,
    stats: &Arc<Mutex<Statistics>>,
    print_opts: print::Options,
) {
    let mut stats = stats.lock().unwrap();

    for line in lines {
        if print_opts.print_metrics {
            println!("metric => {}", line);
        }
        let split = line.split(' ').collect::<Vec<_>>();
        if split.len() != 3 {
            println!("Incorrect graphite metric line: {}", line);
            continue;
        }

        let metric_name = split[0].split('.').last().unwrap().to_string();
        let saved_metric = (*stats).metrics_list.entry(metric_name).or_insert(0);
        *saved_metric += 1;
        (*stats).metrics += 1;

        let metrics_ip_list = (*stats).metrics_ip_list.entry(address).or_insert(0);
        *metrics_ip_list += 1;
    }
}

// Handle metrics in Prometheus format
// Reference: https://help.sumologic.com/Metrics/Introduction-to-Metrics/Metric-Formats#prometheus
pub fn handle_prometheus(
    lines: std::str::Lines,
    address: IpAddr,
    stats: &Arc<Mutex<Statistics>>,
    print_opts: print::Options,
) {
    let mut stats = stats.lock().unwrap();

    for line in lines {
        if print_opts.print_metrics {
            println!("metric => {}", line);
        }
        let metric_name = line.split("{").nth(0).unwrap().to_string();
        let saved_metric = (*stats).metrics_list.entry(metric_name).or_insert(0);
        *saved_metric += 1;
        (*stats).metrics += 1;

        let metrics_ip_list = (*stats).metrics_ip_list.entry(address).or_insert(0);
        *metrics_ip_list += 1;
    }
}
