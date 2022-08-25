use std::collections::HashMap;

pub struct TracesHandleResult {
    pub spans_count: u64,
    pub spans: Vec<Span>,
}

impl TracesHandleResult {
    pub fn new() -> Self {
        TracesHandleResult {
            spans_count: 0,
            spans: vec![],
        }
    }

    pub fn handle_span(&mut self, span: Span) {
        self.spans_count += 1;
        self.spans.push(span);
    }
}

pub type TraceId = String;
pub type SpanId = String;

#[derive(Debug)]
pub struct Span {
    pub name: String,
    pub id: SpanId,
    pub trace_id: TraceId,
    pub parent_span_id: SpanId,
    pub attributes: HashMap<String, String>,
}

impl std::fmt::Display for Span {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        // TODO: Consider adding more info, eg. service name
        write!(
            f,
            "name: {}, span_id: {}, parent_span_id: {}, trace_id: {}",
            self.name,
            self.id,
            self.parent_span_id,
            self.trace_id,
        )
    }
}

pub struct Trace {
    pub span_ids: Vec<SpanId>,
}

impl Trace {
    pub fn new() -> Self {
        Trace { span_ids: vec![] }
    }
}
