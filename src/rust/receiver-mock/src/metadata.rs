use anyhow::anyhow;
use std::collections::HashMap;

pub type Metadata = HashMap<String, String>;

/// Parse the value of the X-Sumo-Fields header into a map of field name to field value
pub fn parse_sumo_fields_header_value(header_value: &str) -> Result<Metadata, anyhow::Error> {
    let mut field_values = HashMap::new();
    if header_value.trim().len() == 0 {
        return Ok(field_values);
    }
    for entry in header_value.split(",") {
        match entry.trim().split_once("=") {
            Some((field_name, field_value)) => field_values.insert(field_name.to_string(), field_value.to_string()),
            None => return Err(anyhow!("Failed to parse X-Sumo-Fields, no `=` in {}", entry)),
        };
    }
    return Ok(field_values);
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_sumo_fields_valid() {
        let single_pair = "_collector=test";
        assert_eq!(
            parse_sumo_fields_header_value(single_pair).unwrap(),
            HashMap::from([(String::from("_collector"), String::from("test"))])
        );

        let multiple_pairs = "service=collection-kube-state-metrics, deployment=collection-kube-state-metrics, node=sumologic-control-plane";
        assert_eq!(
            parse_sumo_fields_header_value(multiple_pairs).unwrap(),
            HashMap::from([
                (
                    String::from("service"),
                    String::from("collection-kube-state-metrics")
                ),
                (
                    String::from("deployment"),
                    String::from("collection-kube-state-metrics")
                ),
                (String::from("node"), String::from("sumologic-control-plane"))
            ])
        );

        let empty = "";
        assert_eq!(parse_sumo_fields_header_value(empty).unwrap(), HashMap::new());
    }

    #[test]
    fn test_parse_sumo_fields_invalid() {
        let invalid_inputs = [",", "no_equals"];
        for input in invalid_inputs {
            assert!(parse_sumo_fields_header_value(input).is_err())
        }
    }
}
