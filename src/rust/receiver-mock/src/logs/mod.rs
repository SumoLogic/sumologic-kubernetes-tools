use crate::time;
use anyhow::Result;
use fancy_regex::Regex;
use log::warn;
use serde_json::Value;
use std::collections::{BTreeMap, HashMap};
use std::net::IpAddr;
use std::sync::{Arc, RwLock};

use crate::metadata::Metadata;

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
    metadata: Metadata,
}


struct RegexCache {
    cache: RwLock<HashMap<String, Arc<Regex>>>,
}

impl RegexCache {
    pub fn new() -> RegexCache {
        RegexCache {
            cache: RwLock::new(HashMap::new()),
        }
    }

    pub fn get(&self, value: &str) -> Result<Arc<Regex>> {
        let map = self.cache.read().unwrap();
        if let Some(regex) = map.get(value) {
            return Ok(regex.clone());
        }
        // Drop the read lock manually to avoid deadlocks when accessing the write lock.
        drop(map);
        let regex = Arc::new(Regex::new(&format!("^{}$", value))?);
        let mut map = self.cache.write().unwrap();
        map.insert(value.to_string(), regex.clone());
        Ok(regex)
    }
}

#[derive(Clone)]
pub struct LogRepository {
    pub messages_by_ts: BTreeMap<u64, Vec<LogMessage>>, // indexed by timestamp to make range queries possible
    regex_cache: Arc<RegexCache>,
}

impl LogRepository {
    pub fn new() -> Self {
        return Self {
            messages_by_ts: BTreeMap::new(),
            regex_cache: Arc::new(RegexCache::new()),
        };
    }

    // This function is a helper to make repository creation in tests easier
    #[cfg(test)]
    pub fn from_raw_logs(raw_logs: Vec<(String, Metadata)>) -> Result<Self, anyhow::Error> {
        let mut repository = Self::new();
        for (body, metadata) in raw_logs {
            repository.add_log_message(body, metadata)
        }
        return Ok(repository);
    }

    pub fn add_log_message(&mut self, body: String, metadata: Metadata) {
        // add the log message to the time index
        let timestamp = match get_timestamp_from_body(&body) {
            Some(ts) => ts,
            None => {
                warn!("Couldn't find timestamp in log line {}", body);
                time::get_now_ms() // use current system time if no timestamp found
            }
        };
        let messages = self.messages_by_ts.entry(timestamp).or_insert(Vec::new());
        messages.push(LogMessage { metadata });
    }

    // Count logs with timestamps in the provided range, with the provided metadata. Empty values
    // in the metadata map mean we just check if the key is there.
    pub fn get_message_count(&self, from_ts: u64, to_ts: u64, metadata_query: HashMap<&str, &str>) -> Result<usize> {
        let mut count = 0;
        let entries = self.messages_by_ts.range(from_ts..to_ts);
        for (_, messages) in entries {
            for message in messages {
                if self.metadata_matches(&metadata_query, &message.metadata)? {
                    count += 1
                }
            }
        }
        return Ok(count);
    }

    // Check if log metadata matches a query in the form of a map of string to string.
    // There's a match if the metadata contains the same keys and values as the query.
    // The query value of an empty string has special meaning, it matches anything.
    fn metadata_matches(&self, query: &HashMap<&str, &str>, target: &Metadata) -> Result<bool> {
        for (key, value) in query.iter() {
            let target_value = match target.get(*key) {
                // get the value from the target
                Some(v) => v,
                None => return Ok(false), // key not present, no match
            };
            if value.len() > 0 {
                // TODO: regex support is available, so we can remove support for "", but it will break the API
                // always match if query value is ""

                // TODO: add caching the regexes
                // To keep backward compatibility, by default match only when whole string is the match.
                let re = self.regex_cache.get(*value)?;

                let match_res = re.is_match(target_value)?;

                if !match_res {
                    return Ok(false);
                }
            }
        }
        return Ok(true);
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
    use std::iter::FromIterator;
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

        repository.add_log_message(body.to_string(), Metadata::new());

        assert_eq!(repository.messages_by_ts.len(), 1);
    }

    #[test]
    fn test_repo_add_message_no_ts() {
        let mut repository = LogRepository::new();
        let body_without_ts = r#"{"log": "Log message"}"#;

        repository.add_log_message(body_without_ts.to_string(), Metadata::new());

        assert_eq!(repository.messages_by_ts.len(), 1);
    }

    #[test]
    fn test_repo_range_query() {
        let timestamps = [1, 5, 8];
        let raw_logs = timestamps
            .iter()
            .map(|ts| format!("{{\"log\": \"Log message\", \"timestamp\": {}}}", ts))
            .map(|body| (body, Metadata::new()))
            .collect();
        let repository = LogRepository::from_raw_logs(raw_logs).unwrap();

        assert_eq!(repository.get_message_count(1, 6, HashMap::new()).unwrap(), 2);
        assert_eq!(repository.get_message_count(0, 10, HashMap::new()).unwrap(), 3);
        assert_eq!(repository.get_message_count(2, 3, HashMap::new()).unwrap(), 0);
    }

    #[test]
    fn test_repo_metadata_query() {
        let metadata = [
            Metadata::new(),
            Metadata::from_iter(vec![("key".to_string(), "value".to_string())].into_iter()),
            Metadata::from_iter(
                vec![
                    ("key".to_string(), "valueprime".to_string()),
                    ("key2".to_string(), "value2".to_string()),
                ]
                .into_iter(),
            ),
        ];
        let body = "{\"log\": \"Log message\", \"timestamp\": 1}";
        let raw_logs = metadata
            .iter()
            .map(|mt| (body.to_string(), mt.to_owned()))
            .collect();
        let repository = LogRepository::from_raw_logs(raw_logs).unwrap();

        assert_eq!(
            repository
                .get_message_count(0, 100, HashMap::from_iter(vec![("key", "value")].into_iter()))
                .unwrap(),
            1
        );
        assert_eq!(
            repository
                .get_message_count(0, 100, HashMap::from_iter(vec![("key", "")].into_iter()))
                .unwrap(),
            2
        );
        assert_eq!(
            repository
                .get_message_count(
                    0,
                    100,
                    HashMap::from_iter(vec![("key", "valueprime"), ("key2", "value2")].into_iter())
                )
                .unwrap(),
            1
        );
    }

    #[test]
    fn test_repo_metadata_query_regex() {
        let metadata = [
            Metadata::from_iter(vec![("key".to_string(), "value".to_string())].into_iter()),
            Metadata::from_iter(vec![("key".to_string(), "valueSUFFIX".to_string())].into_iter()),
            Metadata::from_iter(vec![("key".to_string(), "PREFIXvalue".to_string())].into_iter()),
            Metadata::from_iter(vec![("key".to_string(), "PREFIXvalueSUFFIX".to_string())].into_iter()),
            Metadata::from_iter(vec![("key".to_string(), "undefined".to_string())].into_iter()),
            Metadata::from_iter(vec![("key".to_string(), "not undefined".to_string())].into_iter()),
        ];
        let body = "{\"log\": \"Log message\", \"timestamp\": 1}";
        let raw_logs = metadata
            .iter()
            .map(|mt| (body.to_string(), mt.to_owned()))
            .collect();
        let repository = LogRepository::from_raw_logs(raw_logs).unwrap();

        // Check backward compatibility (match only exact matches)
        assert_eq!(
            repository
                .get_message_count(0, 100, HashMap::from_iter(vec![("key", "value")].into_iter()))
                .unwrap(),
            1
        );

        // Check backward compatibility (empty matches all)
        assert_eq!(
            repository
                .get_message_count(0, 100, HashMap::from_iter(vec![("key", "")].into_iter()))
                .unwrap(),
            6
        );

        assert_eq!(
            repository
                .get_message_count(0, 100, HashMap::from_iter(vec![("key", "value.*")].into_iter()))
                .unwrap(),
            2
        );

        assert_eq!(
            repository
                .get_message_count(
                    0,
                    100,
                    HashMap::from_iter(vec![("key", ".*value.*")].into_iter())
                )
                .unwrap(),
            4
        );

        assert_eq!(
            repository
                .get_message_count(
                    0,
                    100,
                    HashMap::from_iter(vec![("key", "(?!undefined$).*")].into_iter())
                )
                .unwrap(),
            5
        );
    }

    #[test]
    fn test_repo_metadata_regex_cache() {
        let metadata = [
            Metadata::from_iter(
                vec![
                    ("key".to_string(), "value1".to_string()),
                    ("key2".to_string(), "value2".to_string()),
                    ("key3".to_string(), "thirdvalue".to_string()),
                ]
                .into_iter(),
            ),
        ];
        let body = "{\"log\": \"Log message\", \"timestamp\": 1}";
        let raw_logs = metadata
            .iter()
            .map(|mt| (body.to_string(), mt.to_owned()))
            .collect();
        let repository = LogRepository::from_raw_logs(raw_logs).unwrap();

        // No query done yet, confirm that cache is empty.
        assert_eq!(repository.regex_cache.cache.read().unwrap().len(), 0);

        // Do a query
        assert_eq!(
            repository
                .get_message_count(
                    0,
                    100,
                    HashMap::from_iter(vec![("key", "value.*"), ("key3", "third.*")].into_iter())
                )
                .unwrap(),
            1
        );

        // Confirm that cache has increased size
        assert_eq!(repository.regex_cache.cache.read().unwrap().len(), 2);
        
        // Get the regexes.
        let first = repository.regex_cache.cache.read().unwrap().get("value.*").unwrap().clone();
        let third = repository.regex_cache.cache.read().unwrap().get("third.*").unwrap().clone();

        assert_eq!(Arc::strong_count(&first), 2);
        assert_eq!(Arc::strong_count(&third), 2);

        // Do another query
        assert_eq!(
            repository
                .get_message_count(
                    0,
                    100,
                    HashMap::from_iter(vec![("key", "value.*"), ("key2", "value.*"), ("key3", "third.*")].into_iter())
                )
                .unwrap(),
            1
        );

        // Confirm that the cache hasn't changed: the size is the same and the same references are used.
        assert_eq!(repository.regex_cache.cache.read().unwrap().len(), 2);
        assert_eq!(Arc::strong_count(&first), 2);
        assert_eq!(Arc::strong_count(&third), 2);
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
