package fluentdautoscaling

type Fluentd struct {
	Logs    *FluentdLogs           `yaml:"logs,omitempty"`
	Metrics *FluentdMetrics        `yaml:"metrics,omitempty"`
	Rest    map[string]interface{} `yaml:",inline"`
}

type FluentdLogs struct {
	Autoscaling *Autoscaling           `yaml:"autoscaling,omitempty"`
	Rest        map[string]interface{} `yaml:",inline"`
}

type FluentdMetrics = FluentdLogs

type Metadata struct {
	Logs    *MetadataLogs          `yaml:"logs,omitempty"`
	Metrics *MetadataMetrics       `yaml:"metrics,omitempty"`
	Rest    map[string]interface{} `yaml:",inline"`
}

type MetadataLogs struct {
	Autoscaling *Autoscaling           `yaml:"autoscaling,omitempty"`
	Rest        map[string]interface{} `yaml:",inline"`
}

type MetadataMetrics = MetadataLogs

type Autoscaling struct {
	Enabled *bool                  `yaml:"enabled,omitempty"`
	Rest    map[string]interface{} `yaml:",inline"`
}
