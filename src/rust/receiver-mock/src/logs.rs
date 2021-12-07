use anyhow::{anyhow, Error};
use serde_json::Value;
use std::collections::{BTreeMap, HashMap};
use std::net::IpAddr;

#[derive(Clone)]
pub struct LogStats {
    pub count: u64,
    pub bytes: u64,
}

#[derive(Clone)]
pub struct LogMessage {
    // TODO: add body here once there's an API to query it
}

#[derive(Clone)]
pub struct LogRepository {
    pub total: LogStats,
    pub ipaddr_to_stats: HashMap<IpAddr, LogStats>,
    pub messages_by_ts: BTreeMap<u64, Vec<LogMessage>>, // indexed by timestamp to make range queries possible
}

impl LogRepository {
    pub fn new() -> Self {
        return Self {
            total: LogStats { count: 0, bytes: 0 },
            ipaddr_to_stats: HashMap::new(),
            messages_by_ts: BTreeMap::new(),
        };
    }

    pub fn add_log_message(&mut self, body: String, ipaddr: IpAddr) -> Result<(), Error> {
        // add the log message to the time index
        let timestamp = match get_timestamp_from_body(&body) {
            Some(ts) => ts,
            None => return Err(anyhow!("No timestamp found in log message")),
        };
        let messages = self.messages_by_ts.entry(timestamp).or_insert(Vec::new());
        messages.push(LogMessage {});

        // update total stats
        self.total.count += 1;
        self.total.bytes += body.len() as u64;

        // update per ip address stats
        let stats = self
            .ipaddr_to_stats
            .entry(ipaddr)
            .or_insert(LogStats { count: 0, bytes: 0 });
        stats.count += 1;
        stats.bytes += body.len() as u64;

        Ok(())
    }

    pub fn get_message_count(&self, from_ts: u64, to_ts: u64) -> u32 {
        let mut count = 0;
        let entries = self.messages_by_ts.range(from_ts..to_ts);
        for (_, messages) in entries {
            count += messages.len()
        }
        return count as u32;
    }

    pub fn get_stats_for_ipaddr(&self, ipaddr: IpAddr) -> LogStats {
        return self
            .ipaddr_to_stats
            .get(&ipaddr)
            .unwrap_or(&LogStats { count: 0, bytes: 0 })
            .clone();
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

#[cfg(test)]
mod tests {
    // Note this useful idiom: importing names from outer (for mod tests) scope.
    use super::*;
    use std::net::{IpAddr, Ipv4Addr};

    #[test]
    fn test_repo_add_message_valid() {
        let ip_address = IpAddr::V4(Ipv4Addr::new(1, 2, 3, 4));
        let mut repository = LogRepository::new();
        let old_repository = repository.clone();
        let body = r#"{"log": "Log message", "timestamp": 1}"#;

        let result = repository.add_log_message(body.to_string(), ip_address);

        assert!(result.is_ok());
        assert_eq!(repository.total.count, old_repository.total.count + 1);
        assert_eq!(
            repository.total.bytes,
            old_repository.total.bytes + body.len() as u64
        );

        assert_eq!(
            repository.ipaddr_to_stats[&ip_address].count,
            old_repository.get_stats_for_ipaddr(ip_address).count + 1
        );
        assert_eq!(
            repository.ipaddr_to_stats[&ip_address].bytes,
            old_repository.get_stats_for_ipaddr(ip_address).bytes + body.len() as u64
        );
    }

    #[test]
    fn test_repo_add_message_invalid() {
        let ip_address = IpAddr::V4(Ipv4Addr::new(1, 2, 3, 4));
        let mut repository = LogRepository::new();
        let body_without_ts = r#"{"log": "Log message"}"#;

        let result = repository.add_log_message(body_without_ts.to_string(), ip_address);

        assert!(result.is_err());
        assert_eq!(repository.total.count, 0);
    }

    #[test]
    fn test_repo_range_query() {
        let ip_address = IpAddr::V4(Ipv4Addr::new(1, 2, 3, 4));
        let mut repository = LogRepository::new();
        let timestamps = [1, 5, 8];
        let bodies = timestamps
            .iter()
            .map(|ts| format!("{{\"log\": \"Log message\", \"timestamp\": {}}}", ts));

        for body in bodies {
            let result = repository.add_log_message(body.to_string(), ip_address);
            assert!(result.is_ok());
        }

        assert_eq!(repository.total.count, 3);
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
}
