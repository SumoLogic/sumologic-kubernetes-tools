package main

type ValuesV2 struct {
	KubePrometheusStack *KubePrometheusStackV2 `yaml:"kube-prometheus-stack"`
	Rest                map[string]interface{} `yaml:",inline"`
}

type KubeStateMetricsV2 struct {
	Collectors *map[string]bool       `yaml:"collectors"`
	Rest       map[string]interface{} `yaml:",inline"`
}

type KubePrometheusStackV2 struct {
	KubeStateMetrics *KubeStateMetricsV2    `yaml:"kube-state-metrics"`
	Rest             map[string]interface{} `yaml:",inline"`
}
