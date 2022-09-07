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
| `--print-spans`             | `-s`              |      N/A      | Use to print received spans on stdout                                                          |
| `--store-logs`              |                   |      N/A      | Use to store log data which can then be queried via `/logs/*` endpoints                        |
| `--store-metrics`           |                   |      N/A      | Use to store metrics which can then be returned via `/metrics-samples` endpoint                |
| `--store-traces`            |                   |      N/A      | Use to store traces data. Spans can be queried via `/spans-list` endpoint and whole traces can be queries via `/traces-list` endpoint         |
| `--version`                 | `-V`              |      N/A      | Print version information                                                                      |
| `--delay-time` <delay_time> | `-t <delay_time>` |       0       | Use to specify processing delay in milliseconds which will be added to every handled request.      |

## Terraform mock

It expose the `/terraform.*` url which can be used to set HTTP source for k8s collection to receiver-mock itself

Example output:

```json
{"source":{"url":"http://localhost:3333/receiver"}}
```

## Traces

The following endpoints provide information about received traces:

- `/spans-list` - returns list of collected spans

  It accepts a list of key value pairs being an attribute set that the span will
  have to fullfil in order to be returned.
  Attribute values can be omitted in which case only presence of a particular attribute
  will be checked.

  Exemplary output:

  ```shell
  $ curl -s localhost:3000/spans-list\?application=petclinic-app | jq .
  [
  {
    "name": "/**",
    "id": "cb9c07fd1c7c77f7",
    "trace_id": "17b14f4cb48d007be8e169d56ae6a8c5",
    "parent_span_id": "",
    "attributes": {
      "process.runtime.name": "OpenJDK Runtime Environment",
      "process.runtime.description": "Oracle Corporation OpenJDK 64-Bit Server VM 25.272-b10",
      "http.status_code": "404",
      "thread.name": "qtp1687702287-20",
      "http.flavor": "1.1",
      "os.type": "linux",
      "telemetry.sdk.version": "1.11.0",
      "net.peer.port": "60484",
      "application": "petclinic-app",
      "http.scheme": "http",
      "http.route": "/**",
      "process.runtime.version": "1.8.0_272-8u272-b10-0+deb9u1-b10",
      "container.id": "173db7eb967f0673e62e3460bc733fb6d979ec041ac375e75c6c5bdac4d907c9",
      "http.method": "GET",
      "net.transport": "ip_tcp",
      "http.server_name": "localhost",
      "http.user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
      "host.name": "colima",
      "process.executable.path": "/usr/lib/jvm/java-8-openjdk-amd64/jre:bin:java",
      "net.peer.ip": "0:0:0:0:0:0:0:1",
      "service.name": "petclinic-svc",
      "telemetry.auto.version": "1.11.1",
      "telemetry.sdk.language": "java",
      "http.host": "localhost:8080",
      "process.command_line": "/usr/lib/jvm/java-8-openjdk-amd64/jre:bin:java -javaagent:/agent/opentelemetry-javaagent.jar",
      "process.pid": "1",
      "thread.id": "20",
      "host.arch": "amd64",
      "os.description": "Linux 5.10.109-0-virt",
      "telemetry.sdk.name": "opentelemetry",
      "http.target": "/images/spring-logo-dataflow.png"
    }
  },
  # ...
  ]
  ```

  - `/traces-list` - returns list of collected traces

  It accepts a list of key value pairs being an attribute set.
  If any span in a trace contains all of these attributes, this trace will be in the list.
  Attribute values can be omitted in which case only presence of a particular attribute
  will be checked.

  Exemplary output:

  ```shell
  $ curl -s localhost:3000/traces-list\?application=petclinic-app | jq .
  [
    [
      {
        "name": "/resources/**",
        "id": "cf5774ea55249a12",
        "trace_id": "f49cc41476d899cd5372ba38acd3e44b",
        "parent_span_id": "",
        "attributes": {
          "host.name": "colima",
          "http.target": "/resources/images/pets.png",
          "net.peer.port": "51064",
          "process.command_line": "/usr/lib/jvm/java-8-openjdk-amd64/jre:bin:java -javaagent:/agent/opentelemetry-javaagent.jar",
          "thread.name": "qtp1687702287-13",
          "net.peer.ip": "0:0:0:0:0:0:0:1",
          "net.transport": "ip_tcp",
          "http.method": "GET",
          "thread.id": "13",
          "http.route": "/resources/**",
          "process.runtime.description": "Oracle Corporation OpenJDK 64-Bit Server VM 25.272-b10",
          "http.server_name": "localhost",
          "telemetry.auto.version": "1.11.1",
          "telemetry.sdk.language": "java",
          "process.pid": "1",
          "process.runtime.version": "1.8.0_272-8u272-b10-0+deb9u1-b10",
          "process.runtime.name": "OpenJDK Runtime Environment",
          "http.host": "localhost:8080",
          "os.type": "linux",
          "process.executable.path": "/usr/lib/jvm/java-8-openjdk-amd64/jre:bin:java",
          "http.response_content_length": "67721",
          "http.user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36",
          "container.id": "5817c3b8c3815151e0a4dfdb5adf415e383f0aa0cd5ea979f75579c54bca5614",
          "http.flavor": "1.1",
          "application": "petclinic-app",
          "http.status_code": "200",
          "os.description": "Linux 5.10.109-0-virt",
          "host.arch": "amd64",
          "telemetry.sdk.version": "1.11.0",
          "http.scheme": "http",
          "service.name": "petclinic-svc",
          "telemetry.sdk.name": "opentelemetry"
        }
      },
      {
        "name": "ResourceHttpRequestHandler.handleRequest",
        "id": "6640bee2c5f4048a",
        "trace_id": "f49cc41476d899cd5372ba38acd3e44b",
        "parent_span_id": "cf5774ea55249a12",
        "attributes": {
          "process.command_line": "/usr/lib/jvm/java-8-openjdk-amd64/jre:bin:java -javaagent:/agent/opentelemetry-javaagent.jar",
          "process.runtime.name": "OpenJDK Runtime Environment",
          "service.name": "petclinic-svc",
          "telemetry.sdk.version": "1.11.0",
          "os.description": "Linux 5.10.109-0-virt",
          "process.pid": "1",
          "thread.name": "qtp1687702287-13",
          "process.runtime.description": "Oracle Corporation OpenJDK 64-Bit Server VM 25.272-b10",
          "process.executable.path": "/usr/lib/jvm/java-8-openjdk-amd64/jre:bin:java",
          "telemetry.sdk.language": "java",
          "thread.id": "13",
          "host.arch": "amd64",
          "os.type": "linux",
          "telemetry.sdk.name": "opentelemetry",
          "process.runtime.version": "1.8.0_272-8u272-b10-0+deb9u1-b10",
          "telemetry.auto.version": "1.11.1",
          "container.id": "5817c3b8c3815151e0a4dfdb5adf415e383f0aa0cd5ea979f75579c54bca5614",
          "host.name": "colima",
          "application": "petclinic-app"
        }
      }
    ],
    # ...
  ]
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
  Label values can be omitted in which case only presence of a particular label
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
