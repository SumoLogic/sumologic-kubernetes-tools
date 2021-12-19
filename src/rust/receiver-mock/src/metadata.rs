use actix_web::http::header::HeaderMap;
use anyhow::anyhow;
use std::collections::HashMap;

pub type Metadata = HashMap<String, String>;

// Get the metadata from Sumo's common http headers
// This does not include X-Sumo-Fields, which only applies to logs and is handled separately
pub fn get_common_metadata_from_headers(headers: &HeaderMap) -> Result<Metadata, anyhow::Error> {
    let mut metadata = Metadata::new();

    // these metadata field names follow what Sumo itself does
    // see: https://help.sumologic.com/05Search/Get-Started-with-Search/Search-Basics/Built-in-Metadata#built-in-metadata-fields
    // and: https://help.sumologic.com/03Send-Data/Sources/02Sources-for-Hosted-Collectors/HTTP-Source/Upload-Data-to-an-HTTP-Source#supported-http-headers

    match headers.get("x-sumo-name") {
        Some(header_value) => match header_value.to_str() {
            Ok(header_value_str) => metadata.insert("_sourceName".to_string(), header_value_str.to_string()),
            Err(_) => return Err(anyhow!("Couldn't parse X-Sumo-Name header value")),
        },
        None => None,
    };

    match headers.get("x-sumo-host") {
        Some(header_value) => match header_value.to_str() {
            Ok(header_value_str) => metadata.insert("_sourceHost".to_string(), header_value_str.to_string()),
            Err(_) => return Err(anyhow!("Couldn't parse X-Sumo-Host header value")),
        },
        None => None,
    };

    match headers.get("x-sumo-category") {
        Some(header_value) => match header_value.to_str() {
            Ok(header_value_str) => metadata.insert("_sourceCategory".to_string(), header_value_str.to_string()),
            Err(_) => return Err(anyhow!("Couldn't parse X-Sumo-Category header value")),
        },
        None => None,
    };

    return Ok(metadata);
}

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
    use actix_web::http::{HeaderName, HeaderValue};

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

    #[test]
    fn test_get_common_metadata_from_headers_valid() {
        let mut headers = HeaderMap::new();
        headers.insert(
            HeaderName::from_static("x-sumo-name"),
            HeaderValue::from_static("name"),
        );
        headers.insert(
            HeaderName::from_static("x-sumo-host"),
            HeaderValue::from_static("host"),
        );
        headers.insert(
            HeaderName::from_static("x-sumo-category"),
            HeaderValue::from_static("category"),
        );
        let metadata = get_common_metadata_from_headers(&headers).unwrap();

        assert_eq!(
            metadata,
            HashMap::from([
                (String::from("_sourceName"), String::from("name")),
                (String::from("_sourceHost"), String::from("host")),
                (String::from("_sourceCategory"), String::from("category"))
            ])
        )
    }
    #[test]
    fn test_get_common_metadata_from_headers_invalid() {
        let mut headers = HeaderMap::new();
        let invalid_bytes: [u8; 3] = [255, 255, 255];
        headers.insert(
            HeaderName::from_static("x-sumo-name"),
            HeaderValue::from_bytes(&invalid_bytes).unwrap(),
        );
        let result = get_common_metadata_from_headers(&headers);
        match result {
            Ok(_) => panic!("Expected error, got valid result"),
            Err(error) => assert_eq!(error.to_string(), "Couldn't parse X-Sumo-Name header value"),
        }
    }
}
