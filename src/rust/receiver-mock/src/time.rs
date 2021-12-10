use std::time::{SystemTime, UNIX_EPOCH};

pub fn get_now() -> u64 {
    let start = SystemTime::now();
    let since_the_epoch = start.duration_since(UNIX_EPOCH).expect("Time went backwards");
    return since_the_epoch.as_secs() as u64;
}
