package tracingconfig

type ValuesInput struct {
	Otelcol     Otelcol                `yaml:"otelcol,omitempty"`
	Otelgateway Otelgateway            `yaml:"otelgateway,omitempty"`
	Otelagent   map[string]interface{} `yaml:"otelagent,omitempty"`
	Rest        map[string]interface{} `yaml:",inline"`
}

type Otelcol struct {
	Deployment map[string]interface{} `yaml:"deployment,omitempty"`
	Config     struct {
		Processors struct {
			Batch           map[string]interface{} `yaml:"batch,omitempty"`
			CascadingFilter map[string]interface{} `yaml:"cascading_filter,omitempty"`
			MemoryLimiter   map[string]interface{} `yaml:"memory_lmiter,omitempty"`
			Source          map[string]interface{} `yaml:"source,omitempty"`
			Rest            map[string]interface{} `yaml:",inline"`
		} `yaml:"processors,omitempty"`
		Rest map[string]interface{} `yaml:",inline"`
	} `yaml:"config,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}

type Otelgateway struct {
	Deployment map[string]interface{} `yaml:"deployment,omitempty"`
	Config     struct {
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
	Rest map[string]interface{} `yaml:",inline"`
}
