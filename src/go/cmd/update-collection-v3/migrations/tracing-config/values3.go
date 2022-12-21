package tracingconfig

type ValuesV3 struct {
	OtelcolInstrumentation OtelcolInstrumentation `yaml:"otelcolInstrumentation,omitempty"`
	TracesSampler          TracesSampler          `yaml:"tracesSampler,omitempty"`
	Rest                   map[string]interface{} `yaml:",inline"`
}

type OtelcolInstrumentation struct {
	Config struct {
		Processors struct {
			Source map[string]interface{} `yaml:"source,omitempty"`
			Rest   map[string]interface{} `yaml:",inline"`
		} `yaml:"processors,omitempty"`
		Rest map[string]interface{} `yaml:",inline"`
	} `yaml:"config,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}

type TracesSampler struct {
	Config struct {
		Processors struct {
			CascadingFilter map[string]interface{} `yaml:"cascading_filter,omitempty"`
			Rest            map[string]interface{} `yaml:",inline"`
		} `yaml:"processors,omitempty"`
		Rest map[string]interface{} `yaml:",inline"`
	} `yaml:"config,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}
