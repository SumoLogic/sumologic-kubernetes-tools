use std::collections::HashMap;
use std::net::IpAddr;
use std::sync::{Arc, Mutex};

use crate::time::get_now;

pub struct Statistics {
    pub metrics: u64,
    pub logs: u64,
    pub logs_bytes: u64,
    pub p_metrics: u64,
    pub p_logs: u64,
    pub p_logs_bytes: u64,
    pub ts: u64,
    pub metrics_list: HashMap<String, u64>,
    pub metrics_ip_list: HashMap<IpAddr, u64>,
    // logs_ip_list: .0 is logs counter, .1 is bytes counter
    pub logs_ip_list: HashMap<IpAddr, (u64, u64)>,
    pub url: String,
    pub print_logs: bool,
    pub print_headers: bool,
}

pub fn print(stats: &Arc<Mutex<Statistics>>) {
    let mut stats = stats.lock().unwrap();

    if get_now() >= (*stats).ts + 60 {
        println!(
            "{} Metrics: {:10.} Logs: {:10.}; {:6.6} MB/s",
            (*stats).ts,
            (*stats).metrics - (*stats).p_metrics,
            (*stats).logs - (*stats).p_logs,
            (((*stats).logs_bytes - (*stats).p_logs_bytes) as f64)
                / ((get_now() - (*stats).ts) as f64)
                / (1e6 as f64)
        );
        (*stats).ts = get_now();
        (*stats).p_metrics = (*stats).metrics;
        (*stats).p_logs = (*stats).logs;
        (*stats).p_logs_bytes = (*stats).logs_bytes;
    }
}
