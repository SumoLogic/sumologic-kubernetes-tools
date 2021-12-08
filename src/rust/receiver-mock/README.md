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

## Metrics

These are endpoints which provide information about received metrics:

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

- `/metrics-samples` - return metrics samples (last data point for each time series)

  It accepts a list of key value pairs being a label set that the metric sample will
  have to fullfil in order to be returned.
  Label values can be ommitted in which case only presence of a particular label
  will be checked.
  `__name__` is handled specially as it will be matched against the metric name.

  Exemplar output:

  ```shell
  $ curl -s localhost:3000/metrics-samples\?__name__=apiserver_request_total\&cluster | jq .
  [
      {
        "metric": "apiserver_request_total",
        "value": 124,
        "labels": {
          "prometheus_replica": "prometheus-release-test-1638873119-ku-prometheus-0",
          "_origin": "kubernetes",
          "component": "apiserver",
          "service": "kubernetes",
          "resource": "events",
          "code": "422",
          "instance": "172.18.0.2:6443",
          "group": "events.k8s.io",
          "namespace": "default",
          "verb": "POST",
          "scope": "resource",
          "endpoint": "https",
          "version": "v1",
          "job": "apiserver",
          "cluster": "microk8s",
          "prometheus": "ns-test-1638873119/release-test-1638873119-ku-prometheus"
        },
        "timestamp": 163123123
      }
    ]
## Logs

These following endpoints provide information about received logs:

- `/logs/count?from_ts=1&to_ts=1000`

  Returns the number of logs received between `from_ts` and `to_ts`. The values are epoch timestamps in milliseconds, and the range represented by them is inclusive at the start and exclusive at the end. Both values are optional.

  ```json

  {
      "count": 7
  }

  ```

## Disclaimer

This tool is not intended to be used by the 3rd party.
It can significantly change behavior over development time and should be treated as experimental.
