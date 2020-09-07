#[derive(Clone,Copy)]
pub struct Options {
    pub print: Print,
}

#[derive(Clone,Copy)]
pub struct Print {
    pub logs: bool,
    pub headers: bool,
    pub metrics: bool,
}
