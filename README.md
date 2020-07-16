# Sumo Logic Kubernetes Tools
This repository provides set of tools which can be used for debugging and testing [sumologic kubernetes collection](https://github.com/SumoLogic/sumologic-kubernetes-collection/) solution.

# Disclaimer
This toolset is designed for internal usage and it's in development state. We are not giving guarantee of consistency and stability of the application. Inappropriate usage can lead to breaking cluster configuration and/or deployments.

# Running

## K8S Check

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

## Trace stress-tester

There's a simple tool that generates a desired number of spans per minute and sends them using Jaeger format

```
 kubectl run stress-tester \
  -it --rm \
  --restart=Never -n sumologic \
  --image sumologic/kubernetes-tools \
  --serviceaccount='collection-sumologic' \
  --env JAEGER_AGENT_HOST=collection-sumologic-otelcol.sumologic \
  --env JAEGER_AGENT_PORT=6831 \
  --env TOTAL_SPANS=1000000 \
  --env SPANS_PER_MIN=6000 \
  -- stress-tester
```

You can set Jaeger Go client env variables (such as `JAEGER_AGENT_HOST` or `JAEGER_COLLECTOR`) and stress-tester specific ones:

* `TOTAL_SPANS` (default=10000000) - total number of spans to generate
* `SPANS_PER_MIN` (required) - rate of spans per minute (the tester will adjust the delay between iterations to reach such rate)

## Receiver-mock

Small tool for mocking sumologic receiver to avoid sending data outside of cluster.

```bash
$ kubectl run receiver-mock \
 -it --rm \
 --restart=Never \
 --image sumologic/kubernetes-tools \
 -- receiver-mock --help
```

[More information](src/rust/receiver-mock/README.md)

## K8S Template generator

### Generating

Before generating the configuration we recommend to prepare `values.yaml` file where you will store all your configuration.
Alternatively you can replace the file with `--set property=value` arguments according to [helm documentation](https://helm.sh/docs/intro/using_helm/).

#### Docker

```bash
cat values.yaml | docker run \
  --rm -i sumologic/kubernetes-tools \
  template \
    --namespace '<NAMESPACE>' \
    --name-template 'collection' \
      | tee sumologic.yaml
```

#### Kubectl

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

### Applying changes

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

## Interactive mode

The pod can be also run in interactive mode:

```bash
$ kubectl run tools \
  -it --rm \
  --restart=Never \
  --image sumologic/kubernetes-tools \
  -- /bin/bash -l
```
