# Receiver-mock

receiver-mock is an small containerized application written in rust which can be used for local testing of the [`kubernetes sumologic collection`](https://github.com/SumoLogic/sumologic-kubernetes-collection)

## Running

```
cargo run
```

### Arguments

List of arguments taken by receiver-mock:

| Long form                   | Short form        | Default value | Description                                                                                    |
|-----------------------------|-------------------|:-------------:|------------------------------------------------------------------------------------------------|
| `--drop-rate <drop_rate>`   | `-d <drop_rate>`  |       0       | Use to specify packet drop rate. This is number from 0 (do not drop) to 100 (drop all).        |
| `--help`                    | `-h`              |      N/A      | Print help information                                                                         |
| `--hostname <hostname>`     | `-l <hostname>`   |   localhost   | Hostname reported as the receiver. For Kubernetes it will be `<service name>.<namespace name>` |
| `--port <port>`             | `-p <port>`       |     3000      | Port to listen on                                                                              |
| `--print-headers`           |                   |      N/A      | Use to print received request's headers                                                        |
| `--print-logs`              | `-r`              |      N/A      | Use to print received logs on stdout                                                           |
| `--print-metrics`           | `-m`              |      N/A      | Use to print received metrics (with dimensions) on stdout                                      |
| `--store-logs`              |                   |      N/A      | Use to store log data which can then be queried via `/logs/*` endpoints                        |
| `--store-metrics`           |                   |      N/A      | Use to store metrics which can then be returned via `/metrics-samples` endpoint                |
| `--version`                 | `-V`              |      N/A      | Print version information                                                                      |
| `--delay-time` <delay_time> | `-t <delay_time>` |       0       | Use to specify delay time. It mocks request processing time in milliseconds.                   |

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

- `POST /metrics-reset` - reset the metrics counter (zeroes `/metrics-list`)

  Example:

  ```shell
  $ curl -X POST http://localhost:3000/metrics-reset
  All metrics were reset successfully
  ```

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
  ```

## Logs

The following endpoints provide information about received logs:

- `/logs/count?from_ts=1&to_ts=1000&namespace=default&deployment=`

  Returns the number of logs received between `from_ts` and `to_ts`. The values are epoch timestamps in milliseconds, and the range represented by them is inclusive at the start and exclusive at the end. Both values are optional.

  It's also possible to filter by log metadata. Any query parameter without a fixed meaning (such as `from_ts`) will be treated
  as a key-value pair of metadata, and only logs containing that pair will be counted. Similarly to the metrics samples endpoint, an empty value is treated as a wildcard.

  ```json
  {
    "count": 7
  }
  ```

## Dump message

Receiver mock comes with special `/dump` endpoint, which is going to print message on stdout independently on the header value.

## Disclaimer

This tool is not intended to be used by the 3rd party.
It can significantly change behavior over development time and should be treated as experimental.
