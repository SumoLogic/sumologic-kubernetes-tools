#[derive(Clone,Copy)]
pub struct Options {
    pub print_opts: Print,
}

#[derive(Clone,Copy)]
pub struct Print {
    pub print_logs: bool,
    pub print_headers: bool,
    pub print_metrics: bool,
}