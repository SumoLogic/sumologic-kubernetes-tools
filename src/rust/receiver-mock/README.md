# Receiver-mock

receiver-mock is an small contenerised application written in rust which can be used for local testing of the [`kubernetes sumologic collection`](https://github.com/SumoLogic/sumologic-kubernetes-collection)

## Running

```
cargo run
```

### Arguments

List of arguments taken by receiver-mock:

- `-l, --hostname <hostname>`- Hostname reported as the receiver.
  For kubernetes it will be `<service name>.<namespace>` (`localhost` by default)
- `-p, --port <port>` - Port to listen on (default is `3000`)

## Terraform mock

It expose the `/terraform.*` url which can be used to set HTTP source for k8s collection to receiver-mock itself

Example output:

```json
{"source":{"url":"http://localhost:3333/receiver"}}
```

## Statistics

There are endpoints which provides statistics:

- `metrics` - exposes receiver-mock metrics in prometheus format

  ```
  # TYPE receiver_mock_metrics_count counter
  receiver_mock_metrics_count 123
  # TYPE receiver_mock_logs_count counter
  receiver_mock_logs_count 123
  # TYPE receiver_mock_logs_bytes_count counter
  receiver_mock_logs_bytes_count 45678
  ```

- `/metrics-list` - returns list of counted unique metrics

  ```
  prometheus_remote_storage_shards: 100
  prometheus_remote_storage_shards_desired: 100
  prometheus_remote_storage_shards_max: 100
  prometheus_remote_storage_shards_min: 100
  prometheus_remote_storage_string_interner_zero_reference_releases_total: 10
  prometheus_remote_storage_succeeded_samples_total: 100
  ```

- `/metrics-reset` - reset the metrics counter (zeroes `/metrics-list`)

## Disclaimer

This tool is not intended to be used by the 3rd party.
It can significantly change behavior over development time and should be treated as experimental.
