#[derive(Clone, Copy)]
pub struct Options {
    pub print: Print,
    pub drop_rate: i64,
    pub store_metrics: bool,
}

#[derive(Clone, Copy)]
pub struct Print {
    pub logs: bool,
    pub headers: bool,
    pub metrics: bool,
}
