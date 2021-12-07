use crate::time;
use serde_json::Value;
use std::collections::{BTreeMap, HashMap};
use std::net::IpAddr;

#[derive(Clone, Debug, PartialEq)]
pub struct LogStats {
    pub message_count: u64,
    pub byte_count: u64,
}

pub struct LogStatsRepository {
    pub total: LogStats,
    pub ipaddr: HashMap<IpAddr, LogStats>,
}

impl LogStatsRepository {
    pub fn new() -> Self {
        return Self {
            total: LogStats {
                message_count: 0,
                byte_count: 0,
            },
            ipaddr: HashMap::new(),
        };
    }

    pub fn update(&mut self, message_count: u64, byte_count: u64, ipaddr: IpAddr) {
        // update total stats
        self.total.message_count += message_count;
        self.total.byte_count += byte_count;

        // update per ip address stats
        let stats = self.ipaddr.entry(ipaddr).or_insert(LogStats {
            message_count: 0,
            byte_count: 0,
        });
        stats.message_count += message_count;
        stats.byte_count += byte_count as u64;
    }

    #[cfg(test)]
    pub fn get_stats_for_ipaddr(&self, ipaddr: IpAddr) -> LogStats {
        return self
            .ipaddr
            .get(&ipaddr)
            .unwrap_or(&LogStats {
                message_count: 0,
                byte_count: 0,
            })
            .clone();
    }
}

#[derive(Clone)]
pub struct LogMessage {
    // This structure is intended to house more data as we add APIs requiring it
// For example, metadata when we want to query log count by label
}

#[derive(Clone)]
pub struct LogRepository {
    pub messages_by_ts: BTreeMap<u64, Vec<LogMessage>>, // indexed by timestamp to make range queries possible
}

impl LogRepository {
    pub fn new() -> Self {
        return Self {
            messages_by_ts: BTreeMap::new(),
        };
    }

    // This function is a helper to make repository creation in tests easier
    #[cfg(test)]
    pub fn from_raw_logs(raw_logs: Vec<String>) -> Result<Self, anyhow::Error> {
        let mut repository = Self::new();
        for body in raw_logs {
            repository.add_log_message(body)
        }
        return Ok(repository);
    }

    pub fn add_log_message(&mut self, body: String) {
        // add the log message to the time index
        let timestamp = match get_timestamp_from_body(&body) {
            Some(ts) => ts,
            None => {
                eprintln!("Couldn't find timestamp in log line {}", body);
                time::get_now_ms() // use current system time if no timestamp found
            }
        };
        let messages = self.messages_by_ts.entry(timestamp).or_insert(Vec::new());
        messages.push(LogMessage {});
    }

    pub fn get_message_count(&self, from_ts: u64, to_ts: u64) -> usize {
        let mut count = 0;
        let entries = self.messages_by_ts.range(from_ts..to_ts);
        for (_, messages) in entries {
            count += messages.len()
        }
        return count;
    }
}

// Try to get the timestamp from the log body
// We only handle the case where the log is a JSON string representing a map with "timestamp" as a
// top-level key.
fn get_timestamp_from_body(body: &str) -> Option<u64> {
    let parsed_body: Value = match serde_json::from_str(body) {
        Ok(result) => result,
        Err(_) => return None,
    };
    let timestamp = &parsed_body["timestamp"];
    return timestamp.as_u64();
}

/// Parse the value of the X-Sumo-Fields header into a map of field name to field value
fn parse_sumo_fields_header_value(header_value: &str) -> Result<HashMap<String, String>, anyhow::Error> {
    let mut field_values = HashMap::new();
    if header_value.trim().len() == 0 {
        return Ok(field_values);
    }
    for entry in header_value.split(",") {
        match entry.trim().split_once("=") {
            Some((field_name, field_value)) => {
                field_values.insert(field_name.to_string(), field_value.to_string())
            }
            None => {
                return Err(anyhow!(
                    "Failed to parse X-Sumo-Fields, no `=` in {}",
                    entry
                ))
            }
        };
    }
    return Ok(field_values);
}

#[cfg(test)]
mod tests {
    // Note this useful idiom: importing names from outer (for mod tests) scope.
    use super::*;
    use std::net::{IpAddr, Ipv4Addr};

    #[test]
    fn test_stats_repo_update() {
        let ipaddr = IpAddr::V4(Ipv4Addr::new(1, 2, 3, 4));
        let message_count = 5;
        let byte_count = 50;
        let mut repository = LogStatsRepository::new();

        repository.update(message_count, byte_count, ipaddr);

        assert_eq!(repository.total.message_count, message_count);
        assert_eq!(repository.total.byte_count, byte_count);

        assert_eq!(repository.ipaddr[&ipaddr].message_count, message_count);
        assert_eq!(repository.ipaddr[&ipaddr].byte_count, byte_count);

        // check if we get zeroes for an unknown ip address
        let other_ipaddr = IpAddr::V4(Ipv4Addr::new(1, 1, 1, 1));
        assert_eq!(
            repository.get_stats_for_ipaddr(other_ipaddr),
            LogStats {
                message_count: 0,
                byte_count: 0
            }
        )
    }

    #[test]
    fn test_repo_add_message_valid() {
        let mut repository = LogRepository::new();
        let body = r#"{"log": "Log message", "timestamp": 1}"#;

        repository.add_log_message(body.to_string());

        assert_eq!(repository.messages_by_ts.len(), 1);
    }

    #[test]
    fn test_repo_add_message_no_ts() {
        let mut repository = LogRepository::new();
        let body_without_ts = r#"{"log": "Log message"}"#;

        repository.add_log_message(body_without_ts.to_string());

        assert_eq!(repository.messages_by_ts.len(), 1);
    }

    #[test]
    fn test_repo_range_query() {
        let timestamps = [1, 5, 8];
        let raw_logs = timestamps
            .iter()
            .map(|ts| format!("{{\"log\": \"Log message\", \"timestamp\": {}}}", ts))
            .collect();
        let repository = LogRepository::from_raw_logs(raw_logs).unwrap();

        assert_eq!(repository.get_message_count(1, 6), 2);
        assert_eq!(repository.get_message_count(0, 10), 3);
        assert_eq!(repository.get_message_count(2, 3), 0);
    }

    #[test]
    fn test_get_timestamp_from_body() {
        assert_eq!(
            get_timestamp_from_body(r#"{"timestamp": 1234567891011}"#).unwrap(),
            1234567891011
        );
        assert!(get_timestamp_from_body(r#"{"timestamp": -1}"#).is_none());
        assert!(get_timestamp_from_body(r#"{"timestamp": 1.5}"#).is_none());
        assert!(get_timestamp_from_body(r#"{"log": "Some log message"}"#).is_none());
        assert!(get_timestamp_from_body("Not json at all").is_none())
    }

    #[test]
    fn test_parse_sumo_fields_valid() {
        let single_pair = "_collector=test";
        assert_eq!(
            parse_sumo_fields_header_value(single_pair).unwrap(),
            HashMap::from([(String::from("_collector"), String::from("test"))])
        );

        let multiple_pairs = "service=collection-kube-state-metrics, deployment=collection-kube-state-metrics, node=sumologic-control-plane";
        assert_eq!(
            parse_sumo_fields_header_value(multiple_pairs).unwrap(),
            HashMap::from([
                (
                    String::from("service"),
                    String::from("collection-kube-state-metrics")
                ),
                (
                    String::from("deployment"),
                    String::from("collection-kube-state-metrics")
                ),
                (
                    String::from("node"),
                    String::from("sumologic-control-plane")
                )
            ])
        );

        let empty = "";
        assert_eq!(
            parse_sumo_fields_header_value(empty).unwrap(),
            HashMap::new()
        );
    }

    #[test]
    fn test_parse_sumo_fields_invalid() {
        let invalid_inputs = [",", "no_equals"];
        for input in invalid_inputs {
            assert!(parse_sumo_fields_header_value(input).is_err())
        }
    }
}
