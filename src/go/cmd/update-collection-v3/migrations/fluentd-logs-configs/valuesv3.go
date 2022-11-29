package fluentdlogsconfigs

type ValuesV3 struct {
	Sumologic *SumologicV3           `yaml:"sumologic,omitempty"`
	Fluentd   *Fluentd               `yaml:"fluentd,omitempty"`
	Rest      map[string]interface{} `yaml:",inline"`
}

type SumologicV3 struct {
	Logs *SumologicLogsV3       `yaml:"logs,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}

type SumologicLogsV3 struct {
	Container *ContainersLogsConfig  `yaml:"container,omitempty"`
	Systemd   *LogsConfig            `yaml:"systemd,omitempty"`
	Kubelet   *LogsConfig            `yaml:"kubelet,omitempty"`
	Default   *LogsConfig            `yaml:"defaultFluentd,omitempty"`
	Rest      map[string]interface{} `yaml:",inline"`
}
