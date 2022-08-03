# Sumo Logic Kubernetes Tools [![GitHub tag](https://img.shields.io/github/release/SumoLogic/sumologic-kubernetes-tools.svg)](https://gitHub.com/SumoLogic/sumologic-kubernetes-tools/releases/latest)

This repository provides set of tools which can be used for debugging and testing [sumologic kubernetes collection](https://github.com/SumoLogic/sumologic-kubernetes-collection/) solution.

All the various tools are packaged into a single container image that is available in the following public registries:

- Docker Hub [docker.io/sumologic/kubernetes-tools](https://hub.docker.com/r/sumologic/kubernetes-tools/)
- AWS Public ECR [public.ecr.aws/sumologic/kubernetes-tools](https://gallery.ecr.aws/sumologic/kubernetes-tools)

The images are built for the following architectures:

- `linux/amd64`
- `linux/arm64/v8`

## Disclaimer

This toolset is designed for internal usage and it's in development state. We are not giving guarantee of consistency and stability of the application. Inappropriate usage can lead to breaking cluster configuration and/or deployments.

## Requirements

- [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/) >= `1.14`

## Applications

### K8S Check

When Sumo Logic Kubernetes Collection is installed already:

```bash
$ kubectl run tools \
 -it --rm \
 --restart=Never \
 -n sumologic \
 --serviceaccount='collection-sumologic' \
 --image sumologic/kubernetes-tools \
 -- check
```

Alternatively, when collection is not installed, the same command can be run for default serviceaccount:

```bash
$ kubectl run tools \
 -it --rm \
 --restart=Never \
 --image sumologic/kubernetes-tools \
 -- check
```

Should provide an output such as:

```
/var/run/secrets/kubernetes.io/serviceaccount/token exists, size=842
/var/run/secrets/kubernetes.io/serviceaccount/ca.crt exists, size=1025
/var/run/secrets/kubernetes.io/serviceaccount/namespace exists, size=7
/var/run/secrets/kubernetes.io/serviceaccount/namespace contents: default
KUBERNETES_SERVICE_HOST is set
KUBERNETES_SERVICE_PORT is set
POD_NAMESPACE is not set
POD_NAMESPACE env variable:
Kubernetes cluster at 10.96.0.1:443
Running K8S API test
2020/04/21 18:51:45 Kubernetes version: v1.15.5
2020/04/21 18:51:45 Received data for 15 pods in the cluster
pod "diag" deleted
```

### Trace stress-tester

`stress-tester` is a simple tool that generates a desired number of spans per minute and sends them using OpenTelemetry format

```
 kubectl run stress-tester \
  -it --rm \
  --restart=Never -n sumologic \
  --image sumologic/kubernetes-tools \
  --serviceaccount='collection-sumologic' \
  --env COLLECTOR_HOSTNAME=collection-sumologic-otelagent.sumologic \
  --env EXPORTER=http \
  --env TOTAL_SPANS=1000000 \
  --env SPANS_PER_MIN=6000 \
  -- stress-tester
```

#### Configuration

You can set provide configuration to stress-tester by environment variables:

- `TOTAL_SPANS` (default=10000000) - total number of spans to generate
- `SPANS_PER_MIN` (required) - rate of spans per minute (the tester will adjust the delay between iterations to reach such rate)
- `SPANS_PER_TRACE` (default=`100`) - number of spans generated for a single trace
- `COLLECTOR_HOSTNAME` (default=`collection-sumologic-otelagent.sumologic`) - OpenTelemetry collector endpoint 
- `EXPORTER` (default=`http`) - select which exporter is used,  OTLP `http` or `grpc`

### Customer Trace Tester

`customer-trace-tester` is a simple tool that generates a desired number of spans and traces and sends them using OpenTelemetry exporters.
Traces can be easily found with the `service=customer-trace-test-service` filter in the Sumo Logic web application.

```
 kubectl run stress-tester \
  -it --rm \
  --restart=Never -n sumologic \
  --image sumologic/kubernetes-tools \
  --serviceaccount='collection-sumologic' \
  --env COLLECTOR_HOSTNAME=collection-sumologic-otelagent.sumologic \
  --env TOTAL_TRACES=1 \
  --env SPANS_PER_TRACE=10 \
  -- customer-trace-tester
```

#### Configuration

You can configure this tool by setting the following env variables:

- `COLLECTOR_HOSTNAME` (default=`collection-sumologic-otelagent.sumologic`) - the hostname/service of OpenTelemetry Collector
- `TOTAL_TRACES` (default=`1`) - total number of traces to generate
- `SPANS_PER_TRACE` (default=`10`) - number of spans per trace
- `OTLP_HTTP_PORT` (default=`4318`) - port number for OTLP HTTP exporter

#### Example output

```
./customer-trace-tester

2021/07/09 00:32:48 OTLP gRPC Exporter endpoint: collection-sumologic-otelagent.sumologic:4317
2021/07/09 00:32:48 OTLP HTTP Exporter endpoint: collection-sumologic-otelagent.sumologic:4317
2021/07/09 00:32:48 Zipkin Exporter url: http://collection-sumologic-otelagent.sumologic:9411/api/v2/spans
2021/07/09 00:32:48 Jaeger Thrift HTTP Exporter url: http://collection-sumologic-otelagent.sumologic:14268/api/traces
2021/07/09 00:32:48 *******************************
2021/07/09 00:32:48 Sending traces thru otlpHttp exporter
2021/07/09 00:32:48 COLLECTOR_HOSTNAME = collection-sumologic-otelagent.sumologic
2021/07/09 00:32:48 TOTAL_TRACES = 1
2021/07/09 00:32:48 SPANS_PER_TRACE = 10
2021/07/09 00:32:54 *******************************
2021/07/09 00:32:54 Sending traces thru otlpGrpc exporter
2021/07/09 00:32:54 COLLECTOR_HOSTNAME = collection-sumologic-otelagent.sumologic
2021/07/09 00:32:54 TOTAL_TRACES = 1
2021/07/09 00:32:54 SPANS_PER_TRACE = 10
2021/07/09 00:32:59 *******************************
2021/07/09 00:32:59 Sending traces thru zipkin exporter
2021/07/09 00:32:59 COLLECTOR_HOSTNAME = collection-sumologic-otelagent.sumologic
2021/07/09 00:32:59 TOTAL_TRACES = 1
2021/07/09 00:32:59 SPANS_PER_TRACE = 10
2021/07/09 00:33:04 *******************************
2021/07/09 00:33:04 Sending traces thru jaegerThriftHttp exporter
2021/07/09 00:33:04 COLLECTOR_HOSTNAME = collection-sumologic-otelagent.sumologic
2021/07/09 00:33:04 TOTAL_TRACES = 1
2021/07/09 00:33:04 SPANS_PER_TRACE = 10
2021/07/09 00:33:10 *******************************
2021/07/09 00:33:10 Expected number of all traces: 4
2021/07/09 00:33:10 Expected number of spans in single trace: 10
2021/07/09 00:33:10 Expected number of spans for all traces: 40
```

### Receiver-mock

Small tool for mocking sumologic receiver to avoid sending data outside of cluster.

```bash
$ kubectl run receiver-mock \
 -it --rm \
 --restart=Never \
 --image sumologic/kubernetes-tools \
 -- receiver-mock --help
```

[More information](src/rust/receiver-mock/README.md)

### K8S Template generator

#### Generating

Before generating the configuration we recommend to prepare `values.yaml` file where you will store all your configuration.
Alternatively you can replace the file with `--set property=value` arguments according to [helm documentation](https://helm.sh/docs/intro/using_helm/).

##### Docker

```bash
cat values.yaml | docker run \
  --rm -i sumologic/kubernetes-tools \
  template \
    --namespace '<NAMESPACE>' \
    --name-template 'collection' \
      | tee sumologic.yaml
```

##### Kubectl

Minimal supported version of kubectl is `1.14`

```bash
cat values.yaml | \
  kubectl run tools \
    -i --quiet --rm \
    --restart=Never \
    --image sumologic/kubernetes-tools -- \
    template \
      --namespace '<NAMESPACE>' \
      --name-template 'collection' \
      | tee sumologic.yaml
```

#### Applying changes

Due to [issues](https://github.com/helm/charts/tree/master/stable/prometheus-operator#helm-fails-to-create-crds) with prometheus operator and CustomResourceDefinitions you should apply them before applying the generated template.

```
kubectl apply -f https://raw.githubusercontent.com/coreos/prometheus-operator/release-0.38/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagers.yaml
kubectl apply -f https://raw.githubusercontent.com/coreos/prometheus-operator/release-0.38/example/prometheus-operator-crd/monitoring.coreos.com_podmonitors.yaml
kubectl apply -f https://raw.githubusercontent.com/coreos/prometheus-operator/release-0.38/example/prometheus-operator-crd/monitoring.coreos.com_prometheuses.yaml
kubectl apply -f https://raw.githubusercontent.com/coreos/prometheus-operator/release-0.38/example/prometheus-operator-crd/monitoring.coreos.com_prometheusrules.yaml
kubectl apply -f https://raw.githubusercontent.com/coreos/prometheus-operator/release-0.38/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml
kubectl apply -f https://raw.githubusercontent.com/coreos/prometheus-operator/release-0.38/example/prometheus-operator-crd/monitoring.coreos.com_thanosrulers.yaml
```

Wait for CRDs to be created. It should take around few seconds.

Apply the generated template:

```
kubectl apply -f sumologic.yaml
```

### Template dependency configuration

There could be scenarios when you want to get the configuration of the subcharts (prometheus-operator, fluent-bit, etc.).

Command `template-dependency` takes part of the upstream `values.yaml` file basing on the given key:

```
 kubectl run template-dependency \
  -it --quiet --rm \
  --restart=Never -n sumologic \
  --image sumologic/kubernetes-tools \
  -- template-dependency prometheus-operator
```

This command will return our configuration of `prometheus-operator` ready to apply for the `prometheus-operator` helm chart.

You can add additional parameters (like `--version=1.0.0`) at the end of the command.
List of supported arguments is compatible with
[`helm show values`](https://helm.sh/docs/helm/helm_show_values/).

### Kube prometheus mixin configuration

`template-prometheus-mixin` is a command which generates `remoteWrite` mixin configuration for the kube prometheus.

```
 kubectl run template-dependency \
  -it --quiet --rm \
  --restart=Never -n sumologic \
  --image sumologic/kubernetes-tools \
  -- template-prometheus-mixin > kube-prometheus-sumo-logic-mixin.libsonnet
```

You can add additional parameters (like `--version=1.0.0`) at the end of the command.
List of supported arguments is compatible with
[`helm show values`](https://helm.sh/docs/helm/helm_show_values/).

### Logs generator

Logs generator is a tool for generating logs (text lines) using patterns,
which can specify changing parts (words, digits).

```bash
kubectl run template-dependency \
  -it --quiet --rm \
  --restart=Never -n sumologic \
  --image sumologic/kubernetes-tools \
  -- logs-generator --help
```

[More information](src/rust/logs-generator/README.md)

### Interactive mode

The pod can be also run in interactive mode:

```bash
$ kubectl run tools \
  -it --rm \
  --restart=Never \
  --image sumologic/kubernetes-tools \
  -- /bin/bash -l
```
