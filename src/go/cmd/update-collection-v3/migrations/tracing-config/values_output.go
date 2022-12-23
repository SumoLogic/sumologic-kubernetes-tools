package tracingconfig

type ValuesOutput struct {
	OtelcolInstrumentation OtelcolInstrumentation `yaml:"otelcolInstrumentation,omitempty"`
	TracesSampler          TracesSampler          `yaml:"tracesSampler,omitempty"`
	Otelcol                map[string]interface{} `yaml:"-"`
	Otelagent              map[string]interface{} `yaml:"-"`
	Otelgateway            map[string]interface{} `yaml:"-"`
	Rest                   map[string]interface{} `yaml:",inline"`
}

type OtelcolInstrumentation struct {
	Config struct {
		Processors struct {
			Source map[string]interface{} `yaml:"source,omitempty"`
			Rest   map[string]interface{} `yaml:",inline"`
		} `yaml:"processors,omitempty"`
	} `yaml:"config,omitempty"`
}

type TracesSampler struct {
	Config struct {
		Processors struct {
			CascadingFilter map[string]interface{} `yaml:"cascading_filter,omitempty"`
			Rest            map[string]interface{} `yaml:",inline"`
		} `yaml:"processors,omitempty"`
	} `yaml:"config,omitempty"`
}
