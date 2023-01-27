package tracingconfig

type ValuesOutput struct {
	OtelcolInstrumentation OtelcolInstrumentation `yaml:"otelcolInstrumentation,omitempty"`
	TracesGateway          TracesGateway          `yaml:"tracesGateway,omitempty"`
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

type TracesGateway struct {
	Config struct {
		Exporters struct {
			LoadBalancing struct {
				Protocol struct {
					Otlp struct {
						Compression  string `yaml:"compression,omitempty"`
						SendingQueue struct {
							NumConsumers int `yaml:"num_consumers,omitempty"`
							QueueSize    int `yaml:"queue_size,omitempty"`
						} `yaml:"sending_queue,omitempty"`
					} `yaml:"otlp,omitempty"`
				} `yaml:"protocol,omitempty"`
			} `yaml:"loadbalancing,omitempty"`
			Rest map[string]interface{} `yaml:",inline"`
		} `yaml:"exporters,omitempty"`
		Processors struct {
			Batch         map[string]interface{} `yaml:"batch,omitempty"`
			MemoryLimiter map[string]interface{} `yaml:"memory_lmiter,omitempty"`
		} `yaml:"processors,omitempty"`
	} `yaml:"config,omitempty"`
	Deployment map[string]interface{} `yaml:"deployment,omitempty"`
}

type TracesSampler struct {
	Config struct {
		Processors struct {
			Batch           map[string]interface{} `yaml:"batch,omitempty"`
			CascadingFilter map[string]interface{} `yaml:"cascading_filter,omitempty"`
			MemoryLimiter   map[string]interface{} `yaml:"memory_lmiter,omitempty"`
			Rest            map[string]interface{} `yaml:",inline"`
		} `yaml:"processors,omitempty"`
	} `yaml:"config,omitempty"`
	Deployment map[string]interface{} `yaml:"deployment,omitempty"`
}
