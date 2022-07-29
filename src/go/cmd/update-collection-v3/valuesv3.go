package main

type ValuesV3 struct {
	KubePrometheusStack *KubePrometheusStackV3 `yaml:"kube-prometheus-stack,omitempty"`
	Rest                map[string]interface{} `yaml:",inline"`
}

type KubeStateMetricsV3 struct {
	Collectors *[]string              `yaml:"collectors,omitempty"`
	Rest       map[string]interface{} `yaml:",inline"`
}

type KubePrometheusStackV3 struct {
	KubeStateMetrics *KubeStateMetricsV3    `yaml:"kube-state-metrics,omitempty"`
	Rest             map[string]interface{} `yaml:",inline"`
}
