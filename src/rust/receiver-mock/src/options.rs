#[derive(Clone, Copy)]
pub struct Options {
    pub print: Print,
    pub success_ratio: f64,
}

#[derive(Clone, Copy)]
pub struct Print {
    pub logs: bool,
    pub headers: bool,
    pub metrics: bool,
}
