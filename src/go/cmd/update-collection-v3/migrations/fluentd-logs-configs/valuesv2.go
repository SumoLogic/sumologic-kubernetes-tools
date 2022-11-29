package fluentdlogsconfigs

type ValuesV2 struct {
	Sumologic *SumologicV2           `yaml:"sumologic,omitempty"`
	Fluentd   *Fluentd               `yaml:"fluentd,omitempty"`
	Rest      map[string]interface{} `yaml:",inline"`
}

type SumologicV2 struct {
	Logs *SumologicLogsV2       `yaml:"logs,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}

type SumologicLogsV2 struct {
	Container map[string]interface{} `yaml:"container,omitempty"`
	Systemd   map[string]interface{} `yaml:"systemd,omitempty"`
	Rest      map[string]interface{} `yaml:",inline"`
}
