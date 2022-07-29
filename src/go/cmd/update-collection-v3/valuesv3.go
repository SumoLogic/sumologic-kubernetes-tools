package main

type ValuesV3 struct {
	KubePrometheusStack *KubePrometheusStackV3 `yaml:"kube-prometheus-stack"`
	Rest                map[string]interface{} `yaml:",inline"`
}

type KubeStateMetricsV3 struct {
	Collectors *[]string              `yaml:"collectors"`
	Rest       map[string]interface{} `yaml:",inline"`
}

type KubePrometheusStackV3 struct {
	KubeStateMetrics *KubeStateMetricsV3    `yaml:"kube-state-metrics"`
	Rest             map[string]interface{} `yaml:",inline"`
}
