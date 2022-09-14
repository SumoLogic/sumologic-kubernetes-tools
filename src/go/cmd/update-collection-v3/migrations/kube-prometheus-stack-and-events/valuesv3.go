package kubeprometheusstackandevents

type ValuesV3 struct {
	Sumologic struct {
		Events CommonEventsV3         `yaml:"events,omitempty"`
		Rest   map[string]interface{} `yaml:",inline"`
	} `yaml:"sumologic,omitempty"`
	Fluentd struct {
		Events      FluentDEventsV3        `yaml:"events,omitempty"`
		Persistence FluentDPersistenceV2   `yaml:"persistence,omitempty"`
		Rest        map[string]interface{} `yaml:",inline"`
	} `yaml:"fluentd,omitempty"`
	Otelevents          OteleventsV3           `yaml:"otelevents,omitempty"`
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

type FluentDEventsV3 struct{}

type CommonEventsV3 struct {
	Enabled        *bool   `yaml:"enabled,omitempty"`
	Provider       *string `yaml:"provider,omitempty"`
	SourceName     *string `yaml:"sourceName,omitempty"`
	SourceCategory *string `yaml:"sourceCategory,omitempty"`
	Persistence    struct {
		Enabled          *bool   `yaml:"enabled,omitempty"`
		Size             *string `yaml:"size,omitempty"`
		PersistentVolume struct {
			Path         *string                `yaml:"path,omitempty"`
			AccessMode   *string                `yaml:"accessMode,omitempty"`
			StorageClass *string                `yaml:"storageClass,omitempty"`
			PvcLabels    map[string]interface{} `yaml:"pvcLabels,omitempty"`
		} `yaml:"persistentVolume,omitempty"`
	} `yaml:"persistence,omitempty"`
}

type OteleventsV3 struct {
	Rest map[string]interface{} `yaml:",inline"`
}
