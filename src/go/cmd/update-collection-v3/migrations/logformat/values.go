package logformat

type Values struct {
	Sumologic *Sumologic             `yaml:"sumologic,omitempty"`
	Fluentd   *Fluentd               `yaml:"fluentd,omitempty"`
	Rest      map[string]interface{} `yaml:",inline"`
}

type Sumologic struct {
	Logs *SumologicLogs         `yaml:"logs,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}

type SumologicLogs struct {
	Container *SumologicLogsContainer `yaml:"container,omitempty"`
	Rest      map[string]interface{}  `yaml:",inline"`
}

type SumologicLogsContainer struct {
	Format *string                `yaml:"format,omitempty"`
	Rest   map[string]interface{} `yaml:",inline"`
}

type Fluentd struct {
	Logs *FluentdLogs           `yaml:"logs,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}

type FluentdLogs struct {
	Output *FluentdLogsOutput     `yaml:"output,omitempty"`
	Rest   map[string]interface{} `yaml:",inline"`
}

type FluentdLogsOutput struct {
	LogFormat *string                `yaml:"logFormat,omitempty"`
	Rest      map[string]interface{} `yaml:",inline"`
}

type InputValues = Values
type OutputValues = Values
