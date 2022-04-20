use std::time;

#[derive(Clone, Copy)]
pub struct Options {
    pub print: Print,
    pub drop_rate: i64,
    pub delay_time: time::Duration,
    pub store_metrics: bool,
    pub store_logs: bool,
}

#[derive(Clone, Copy)]
pub struct Print {
    pub logs: bool,
    pub headers: bool,
    pub metrics: bool,
}
