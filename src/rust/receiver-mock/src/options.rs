#[derive(Clone, Copy)]
pub struct Options {
    pub print: Print,
    pub success_ratio: f64,
    pub min_wait_time: u64,
    pub max_wait_time: u64,
}

#[derive(Clone, Copy)]
pub struct Print {
    pub logs: bool,
    pub headers: bool,
    pub metrics: bool,
}
