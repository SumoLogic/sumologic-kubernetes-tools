use std::collections::{HashMap, HashSet};
use std::hash::{Hash, Hasher};

use serde::{Deserialize, Serialize};

// Would love to use predefined structs from prometheus_parse create but since those
// don't define Serialize/Deserialize impls we can't.
// ref: https://serde.rs/remote-derive.html
#[derive(Debug, Deserialize, Serialize, Clone)]
// #[derive(PartialEq)]
pub struct Sample {
    pub metric: String,
    pub value: f64,
    pub labels: HashMap<String, String>,
    pub timestamp: u64, // milliseconds epoch timestamp
}

impl PartialEq for Sample {
    fn eq(&self, other: &Self) -> bool {
        self.metric == other.metric && self.labels.eq(&other.labels)
    }
}

impl Eq for Sample {}

impl Hash for Sample {
    fn hash<H: Hasher>(&self, state: &mut H) {
        self.metric.hash(state);

        // Sort otherwise we get a different hash
        let mut sorted_labels: Vec<_> = self.labels.iter().collect();
        sorted_labels.sort_by(|x, y| x.0.cmp(&y.0));

        for (name, value) in sorted_labels {
            name.hash(state);
            value.hash(state);
            // This would produce a different hash :()
            // self.labels.hash(state);
        }
    }
}

pub fn lines_to_samples(lines: Vec<String>) -> HashSet<Sample> {
    let scrape = prometheus_parse::Scrape::parse(lines.into_iter().map(|s| Ok(s))).unwrap();
    let samples = scrape.samples;

    samples
        .iter()
        .map(|sample| {
            let n = sample.labels.len();
            let mut labels: HashMap<String, String> = HashMap::with_capacity(n);
            for s in sample.labels.iter() {
                labels.insert(s.0.to_owned(), s.1.to_owned());
            }

            let value = match sample.value {
                prometheus_parse::Value::Counter(v)
                | prometheus_parse::Value::Gauge(v)
                | prometheus_parse::Value::Untyped(v) => v,
                // Don't support summaries and histograms
                _ => 0.0,
            };

            Sample {
                metric: sample.metric.clone(),
                value: value,
                labels: labels,
                timestamp: sample.timestamp.timestamp_millis() as u64,
            }
        })
        .collect()
}

// filter_samples filters the provided samples based on the provided labels.
// In order for the sample to be returned it has to contain all the provided labels
// with provided values. If a label value is not provided then it's checked for
// existence within the sample.
// `__name__` is handled specially as it will be matched against the metric name.
pub fn filter_samples(samples: &HashSet<Sample>, labels: HashMap<String, String>) -> HashSet<Sample> {
    samples
        .iter()
        .filter(|sample| {
            // For every provided param 'key-value' pair...
            for (param_key, param_val) in &labels {
                // In order to keep the params simply a key value list let's treat
                // '__name__' specially so that it matches the metric name.
                if param_key == "__name__" {
                    if param_val != "" && &sample.metric != param_val {
                        // If the metric name doesn't match the provided '__name__'
                        // value then drop the sample.
                        return false;
                    }
                    // Otherwise continue (get next key value pair from params)
                    continue;
                }

                // ...try to find it in sample's labels...
                match sample.labels.get(&param_key[..]) {
                    Some(sample_value) => {
                        // ...if sample contains it and query param was provided
                        // without a value then keep iterating...
                        if param_val == "" {
                            continue;
                        }

                        // ...if the value was provided and it matches sample's
                        // label value then also keep iterating...
                        if sample_value == param_val {
                            continue;
                        }

                        // ...otherwise drop this sample: the requested label has
                        // a different value.
                        return false;
                    }

                    // If the requested label wasn't found in sample's labels then bail.
                    None => return false,
                }
            }
            true
        })
        .cloned()
        .collect()
}
