use std::collections::HashMap;
use std::net::IpAddr;

#[derive(Clone)]
pub struct LogStats {
    pub count: u64,
    pub bytes: u64,
}

#[derive(Clone)]
pub struct LogRepository {
    pub total: LogStats,
    pub ipaddr_to_stats: HashMap<IpAddr, LogStats>,
}

impl LogRepository {
    pub fn new() -> Self {
        return Self {
            total: LogStats { count: 0, bytes: 0 },
            ipaddr_to_stats: HashMap::new(),
        };
    }

    pub fn add_log_message(&mut self, body: String, ipaddr: IpAddr) {
        self.total.count += 1;
        self.total.bytes += body.len() as u64;
        let stats = self
            .ipaddr_to_stats
            .entry(ipaddr)
            .or_insert(LogStats { count: 0, bytes: 0 });
        stats.count += 1;
        stats.bytes += body.len() as u64;
    }

    pub fn get_stats_for_ipaddr(&self, ipaddr: IpAddr) -> LogStats {
        return self
            .ipaddr_to_stats
            .get(&ipaddr)
            .unwrap_or(&LogStats { count: 0, bytes: 0 })
            .clone();
    }
}

#[cfg(test)]
mod tests {
    // Note this useful idiom: importing names from outer (for mod tests) scope.
    use super::*;
    use std::net::{IpAddr, Ipv4Addr};

    #[test]
    fn test_add_message() {
        let ip_address = IpAddr::V4(Ipv4Addr::new(1, 2, 3, 4));
        let mut repository = LogRepository::new();
        let old_repository = repository.clone();
        let body = "Log message";

        repository.add_log_message(body.to_string(), ip_address);

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
}
