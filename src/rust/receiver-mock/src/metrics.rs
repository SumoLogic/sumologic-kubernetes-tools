use std::collections::HashMap;
use std::net::IpAddr;

use crate::options;

pub struct MetricsHandleResult {
    pub metrics: u64,
    pub metrics_list: HashMap<String, u64>,
    pub metrics_ip_list: HashMap<IpAddr, u64>,
}

impl MetricsHandleResult {
    fn handle_metric(&mut self, metric_name: String) {
        let saved_metric = self.metrics_list.entry(metric_name).or_insert(0);
        *saved_metric += 1;
        self.metrics += 1;
    }

    fn handle_ip(&mut self, ip_address: IpAddr) {
        let metrics_ip_list = self.metrics_ip_list.entry(ip_address).or_insert(0);
        *metrics_ip_list += 1;
    }
}

// Handle metrics in Carbon2.0 format
// Reference: https://help.sumologic.com/Metrics/Introduction-to-Metrics/Metric-Formats#carbon-2-0
pub fn handle_carbon2(
    lines: std::str::Lines,
    address: IpAddr,
    print_opts: options::Print,
) -> MetricsHandleResult {
    let mut result = MetricsHandleResult {
        metrics: 0,
        metrics_list: HashMap::new(),
        metrics_ip_list: HashMap::new(),
    };

    for line in lines {
        if print_opts.metrics {
            println!("metric => {}", line);
        }
        let mut split = line.split("  ");
        let intrinsic_metrics = split.nth(0).unwrap();
        for metric in intrinsic_metrics.split(" ") {
            let metric_name = metric.split("=").nth(0).unwrap().to_string();
            if metric_name == "metric" {
                result.handle_metric(metric_name);
                break;
            }
        }

        result.handle_ip(address);
    }

    result
}

// Handle metrics in Graphite format
// Reference: https://help.sumologic.com/Metrics/Introduction-to-Metrics/Metric-Formats#graphite
pub fn handle_graphite(
    lines: std::str::Lines,
    address: IpAddr,
    print_opts: options::Print,
) -> MetricsHandleResult {
    let mut result = MetricsHandleResult {
        metrics: 0,
        metrics_list: HashMap::new(),
        metrics_ip_list: HashMap::new(),
    };

    for line in lines {
        if print_opts.metrics {
            println!("metric => {}", line);
        }
        let split = line.split(' ').collect::<Vec<_>>();
        if split.len() != 3 {
            println!("Incorrect graphite metric line: {}", line);
            continue;
        }

        let metric_name = split[0].split('.').last().unwrap().to_string();
        result.handle_metric(metric_name);
        result.handle_ip(address);
    }

    result
}

// Handle metrics in Prometheus format
// Reference: https://help.sumologic.com/Metrics/Introduction-to-Metrics/Metric-Formats#prometheus
pub fn handle_prometheus(
    lines: std::str::Lines,
    address: IpAddr,
    print_opts: options::Print,
) -> MetricsHandleResult {
    let mut result = MetricsHandleResult {
        metrics: 0,
        metrics_list: HashMap::new(),
        metrics_ip_list: HashMap::new(),
    };

    for line in lines {
        // Ignore comments
        if line.starts_with("#") {
            continue
        }

        if print_opts.metrics {
            println!("metric => {}", line);
        }
        let metric_name = line.split("{").nth(0).unwrap().to_string();
        result.handle_metric(metric_name);
        result.handle_ip(address);
    }

    result
}
