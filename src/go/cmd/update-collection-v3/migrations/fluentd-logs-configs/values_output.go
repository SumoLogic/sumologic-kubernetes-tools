package fluentdlogsconfigs

type ValuesOutput struct {
	Sumologic *SumologicOutput       `yaml:"sumologic,omitempty"`
	Fluentd   *Fluentd               `yaml:"fluentd,omitempty"`
	Rest      map[string]interface{} `yaml:",inline"`
}

type SumologicOutput struct {
	Logs *SumologicLogsOutput   `yaml:"logs,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}

type SumologicLogsOutput struct {
	Container *ContainersLogsConfig  `yaml:"container,omitempty"`
	Systemd   *LogsConfig            `yaml:"systemd,omitempty"`
	Kubelet   *LogsConfig            `yaml:"kubelet,omitempty"`
	Default   *LogsConfig            `yaml:"defaultFluentd,omitempty"`
	Rest      map[string]interface{} `yaml:",inline"`
}
