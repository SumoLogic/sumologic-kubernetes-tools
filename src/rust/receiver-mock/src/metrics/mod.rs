use log::debug;
use std::collections::HashMap;
use std::collections::HashSet;
use std::net::IpAddr;

use sample::*;

use crate::options;

pub mod sample;

pub struct MetricsHandleResult {
    pub metrics_count: u64,
    pub metrics_list: HashMap<String, u64>,
    pub metrics_ip_list: HashMap<IpAddr, u64>,
    pub metrics_samples: HashSet<Sample>,
}

impl MetricsHandleResult {
    pub fn new() -> Self {
        return Self {
            metrics_count: 0,
            metrics_list: HashMap::new(),
            metrics_ip_list: HashMap::new(),
            metrics_samples: HashSet::new(),
        };
    }

    pub fn handle_metric(&mut self, metric_name: String) {
        let saved_metric = self.metrics_list.entry(metric_name).or_insert(0);
        *saved_metric += 1;
        self.metrics_count += 1;
    }

    pub fn handle_ip(&mut self, ip_address: IpAddr) {
        let metrics_ip_list = self.metrics_ip_list.entry(ip_address).or_insert(0);
        *metrics_ip_list += 1;
    }
}

// Handle metrics in Carbon2.0 format
// Reference: https://help.sumologic.com/Metrics/Introduction-to-Metrics/Metric-Formats#carbon-2-0
pub fn handle_carbon2(lines: std::str::Lines, address: IpAddr, print_opts: options::Print) -> MetricsHandleResult {
    let mut result = MetricsHandleResult::new();

    for line in lines {
        if print_opts.metrics {
            debug!("metric => {}", line);
        }
        let mut split = line.split("  ");
        let intrinsic_metrics = split.nth(0).unwrap();
        for metric in intrinsic_metrics.split(" ") {
            let mut split = metric.split("=");
            let field_name = split.nth(0).unwrap().to_string();
            if field_name == "metric" {
                // nth() consumes elements hence nth(0) again
                let metric_name = split.nth(0).unwrap().to_string();
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
pub fn handle_graphite(lines: std::str::Lines, address: IpAddr, print_opts: options::Print) -> MetricsHandleResult {
    let mut result = MetricsHandleResult::new();

    for line in lines {
        if print_opts.metrics {
            debug!("metric => {}", line);
        }
        let split_line = line.split(' ').collect::<Vec<_>>();
        if split_line.len() != 3 {
            debug!("Incorrect graphite metric line: {}", line);
            continue;
        }

        let split_metric = split_line[0].split('.').collect::<Vec<_>>();
        if split_metric.len() != 3 {
            debug!("Incorrect graphite metric name: {}", split_line[0]);
            continue;
        }

        let metric_name = split_metric[1];
        let metric_field = split_metric[2];
        result.handle_metric(format!("{}_{}", metric_name, metric_field));
        result.handle_ip(address);
    }

    result
}

// Handle metrics in Prometheus format
// Reference: https://help.sumologic.com/Metrics/Introduction-to-Metrics/Metric-Formats#prometheus
pub fn handle_prometheus(lines: std::str::Lines, address: IpAddr, opts: &options::Options) -> MetricsHandleResult {
    let mut result = MetricsHandleResult::new();

    let mut lines_vec = vec![];
    for l in lines {
        if l.starts_with("#") {
            continue;
        }

        if opts.print.metrics {
            debug!("metric => {}", l);
        }
        // This should also be implemented in terms of parsed metrics, see below.
        let metric_name = l.split("{").nth(0).unwrap().to_string();
        result.handle_metric(metric_name);
        result.handle_ip(address);

        if opts.store_metrics {
            lines_vec.push(l.to_owned());
        }
    }

    if opts.store_metrics {
        result.metrics_samples = lines_to_samples(lines_vec);
    }

    result
}

#[cfg(test)]
mod tests {
    // Note this useful idiom: importing names from outer (for mod tests) scope.
    use super::*;
    use std::net::{IpAddr, Ipv4Addr};

    #[test]
    fn test_carbon_basic() {
        let lines = "metric=mem_available_percent host=myhostname  50.430792570114136 1601906858
metric=mem_free host=myhostname  12677414912 1601906858
metric=mem_total host=myhostname  68719476736 1601906858
metric=mem_used host=myhostname  34063699968 1601906858
metric=mem_used_percent host=myhostname  49.569207429885864 1601906858
metric=mem_active host=myhostname  25058705408 1601906858
metric=mem_inactive host=myhostname  21978361856 1601906858
metric=mem_wired host=myhostname  5661790208 1601906858
metric=mem_available host=myhostname  34655776768 1601906858"
            .lines();

        let ip_address = IpAddr::V4(Ipv4Addr::new(1, 2, 3, 4));
        let print_opts = options::Print {
            logs: false,
            headers: false,
            metrics: false,
            spans: false,
        };
        let result = handle_carbon2(lines, ip_address, print_opts);

        assert_eq!(result.metrics_count, 9);

        let mut metrics_list: HashMap<String, u64> = HashMap::new();
        metrics_list.insert(String::from("mem_available_percent"), 1);
        metrics_list.insert(String::from("mem_free"), 1);
        metrics_list.insert(String::from("mem_total"), 1);
        metrics_list.insert(String::from("mem_used"), 1);
        metrics_list.insert(String::from("mem_used_percent"), 1);
        metrics_list.insert(String::from("mem_active"), 1);
        metrics_list.insert(String::from("mem_inactive"), 1);
        metrics_list.insert(String::from("mem_wired"), 1);
        metrics_list.insert(String::from("mem_available"), 1);

        assert_eq!(result.metrics_list, metrics_list);

        assert_eq!(result.metrics_ip_list.contains_key(&ip_address), true);
        assert_eq!(*result.metrics_ip_list.get(&ip_address).unwrap(), 9);
    }

    #[test]
    fn test_prometheus_basic() {
        let lines = r##"mem_available_percent{host="myhostname"} 49.59816932678223
mem_active{host="myhostname"} 2.56055296e+10
mem_inactive{host="myhostname"} 2.2181629952e+10
mem_wired{host="myhostname"} 5.678206976e+09
mem_total{host="myhostname"} 6.8719476736e+10
mem_available{host="myhostname"} 3.4083602432e+10
mem_used{host="myhostname"} 3.4635874304e+10
mem_used_percent{host="myhostname"} 50.40183067321777
mem_free{host="myhostname"} 1.190197248e+10"##
            .lines();

        let ip_address = IpAddr::V4(Ipv4Addr::new(1, 2, 3, 4));
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
            store_logs: true,
        };
        let result = handle_prometheus(lines, ip_address, &opts);

        assert_eq!(result.metrics_count, 9);

        let mut metrics_list: HashMap<String, u64> = HashMap::new();
        metrics_list.insert(String::from("mem_available_percent"), 1);
        metrics_list.insert(String::from("mem_free"), 1);
        metrics_list.insert(String::from("mem_total"), 1);
        metrics_list.insert(String::from("mem_used"), 1);
        metrics_list.insert(String::from("mem_used_percent"), 1);
        metrics_list.insert(String::from("mem_active"), 1);
        metrics_list.insert(String::from("mem_inactive"), 1);
        metrics_list.insert(String::from("mem_wired"), 1);
        metrics_list.insert(String::from("mem_available"), 1);

        assert_eq!(result.metrics_list, metrics_list);

        assert_eq!(result.metrics_ip_list.contains_key(&ip_address), true);
        assert_eq!(*result.metrics_ip_list.get(&ip_address).unwrap(), 9);
    }

    #[test]
    fn test_graphite_basic() {
        let lines = "myhostname.mem.available 33310904320 1601909210
myhostname.mem.used_percent 51.526254415512085 1601909210
myhostname.mem.available_percent 48.473745584487915 1601909210
myhostname.mem.active 26373685248 1601909210
myhostname.mem.total 68719476736 1601909210
myhostname.mem.used 35408572416 1601909210
myhostname.mem.inactive 22692282368 1601909210
myhostname.mem.free 10618621952 1601909210
myhostname.mem.wired 5680394240 1601909210"
            .lines();

        let ip_address = IpAddr::V4(Ipv4Addr::new(1, 2, 3, 4));
        let print_opts = options::Print {
            logs: false,
            headers: false,
            metrics: false,
            spans: false,
        };
        let result = handle_graphite(lines, ip_address, print_opts);

        assert_eq!(result.metrics_count, 9);

        let mut metrics_list: HashMap<String, u64> = HashMap::new();
        metrics_list.insert(String::from("mem_available_percent"), 1);
        metrics_list.insert(String::from("mem_free"), 1);
        metrics_list.insert(String::from("mem_total"), 1);
        metrics_list.insert(String::from("mem_used"), 1);
        metrics_list.insert(String::from("mem_used_percent"), 1);
        metrics_list.insert(String::from("mem_active"), 1);
        metrics_list.insert(String::from("mem_inactive"), 1);
        metrics_list.insert(String::from("mem_wired"), 1);
        metrics_list.insert(String::from("mem_available"), 1);

        assert_eq!(result.metrics_list, metrics_list);

        assert_eq!(result.metrics_ip_list.contains_key(&ip_address), true);
        assert_eq!(*result.metrics_ip_list.get(&ip_address).unwrap(), 9);
    }
}
