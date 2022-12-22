package tracingreplaces

type ValuesV2 struct {
	Otelcol     Otelcol                `yaml:"otelcol,omitempty"`
	Otelagent   map[string]interface{} `yaml:"otelagent,omitempty"`
	Otelgateway map[string]interface{} `yaml:"otelgateway,omitempty"`
	Rest        map[string]interface{} `yaml:",inline"`
}

type Otelcol struct {
	Config struct {
		Processors struct {
			CascadingFilter map[string]interface{} `yaml:"cascading_filter,omitempty"`
			Source          map[string]interface{} `yaml:"source,omitempty"`
			Rest            map[string]interface{} `yaml:",inline"`
		} `yaml:"processors,omitempty"`
		Rest map[string]interface{} `yaml:",inline"`
	} `yaml:"config,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}
