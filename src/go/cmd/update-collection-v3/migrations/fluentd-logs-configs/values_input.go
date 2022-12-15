package fluentdlogsconfigs

type ValuesInput struct {
	Sumologic *SumologicInput        `yaml:"sumologic,omitempty"`
	Fluentd   *Fluentd               `yaml:"fluentd,omitempty"`
	Rest      map[string]interface{} `yaml:",inline"`
}

type SumologicInput struct {
	Logs *SumologicLogsInput    `yaml:"logs,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}

type SumologicLogsInput struct {
	Container map[string]interface{} `yaml:"container,omitempty"`
	Systemd   map[string]interface{} `yaml:"systemd,omitempty"`
	Rest      map[string]interface{} `yaml:",inline"`
}
