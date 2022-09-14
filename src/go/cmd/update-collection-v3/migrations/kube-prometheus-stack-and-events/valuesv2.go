package kubeprometheusstackandevents

type ValuesV2 struct {
	Sumologic struct {
		Events CommonEventsV2         `yaml:"events,omitempty"`
		Rest   map[string]interface{} `yaml:",inline"`
	} `yaml:"sumologic,omitempty"`
	Fluentd struct {
		Events      FluentDEventsV2        `yaml:"events,omitempty"`
		Persistence FluentDPersistenceV2   `yaml:"persistence,omitempty"`
		Rest        map[string]interface{} `yaml:",inline"`
	} `yaml:"fluentd,omitempty"`
	Otelevents          OteleventsV2           `yaml:"otelevents,omitempty"`
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

type FluentDEventsV2 struct {
	Enabled        *bool   `yaml:"enabled,omitempty"`
	SourceName     *string `yaml:"sourceName,omitempty"`
	SourceCategory *string `yaml:"sourceCategory,omitempty"`
}

type FluentDPersistenceV2 struct {
	Enabled      *bool   `yaml:"enabled,omitempty"`
	Size         *string `yaml:"size,omitempty"`
	AccessMode   *string `yaml:"accessMode,omitempty"`
	StorageClass *string `yaml:"storageClass"`
}

type CommonEventsV2 struct {
	Enabled  *bool   `yaml:"enabled,omitempty"`
	Provider *string `yaml:"provider,omitempty"`
}

type OteleventsV2 struct {
	Persistence struct {
		Enabled    *bool                  `yaml:"enabled,omitempty"`
		AccessMode *string                `yaml:"accessMode,omitempty"`
		Size       *string                `yaml:"size,omitempty"`
		PvcLabels  map[string]interface{} `yaml:"pvcLabels,omitempty"`
	} `yaml:"persistence,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}
