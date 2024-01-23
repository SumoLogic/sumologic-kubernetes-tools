use std::collections::HashMap;

use log::warn;
use serde::{Deserialize, Serialize};

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

#[derive(Debug, Default, Deserialize, Serialize)]
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
            self.name, self.id, self.parent_span_id, self.trace_id,
        )
    }
}

fn is_span_ok(span: &Span, params: &HashMap<String, String>) -> bool {
    for (key, value) in params.iter() {
        // Identically as in the metric's case,
        // we use "__name__" as key for span's name
        // to keep the querying simple.
        if key == "__name__" {
            if value.is_empty() || span.name == *value {
                continue;
            }
            return false;
        }

        if let Some(val) = span.attributes.get(key) {
            if val.eq(value) {
                continue;
            }

            return false;
        } else {
            return false;
        }
    }

    true
}

pub fn filter_spans<'a>(spans: impl Iterator<Item = &'a Span>, params: HashMap<String, String>) -> Vec<&'a Span> {
    spans.filter(|span| is_span_ok(span, &params)).collect()
}

pub fn filter_traces<'a>(
    traces: impl Iterator<Item = &'a Trace>,
    spans: &'a HashMap<SpanId, Span>,
    params: HashMap<String, String>,
) -> Vec<Vec<&'a Span>> {
    traces
        .map(|trace| {
            // Doing this functionally would be a mess if we want to handle bugs without panicking.
            let mut spans_vec = Vec::with_capacity(trace.span_ids.len());
            for span_id in &trace.span_ids {
                if let Some(span) = spans.get(span_id) {
                    spans_vec.push(span);
                } else {
                    warn!("Span with id {} not found", span_id);
                }
            }
            spans_vec
        })
        .filter(|spans_vec| spans_vec.iter().any(|&span| is_span_ok(span, &params)))
        .collect()
}

pub struct Trace {
    pub span_ids: Vec<SpanId>,
}

impl Trace {
    pub fn new() -> Self {
        Trace { span_ids: vec![] }
    }
}
