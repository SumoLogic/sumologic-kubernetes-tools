package tracingreplaces

type ValuesV2 struct {
	Otelagent *Otelagent `yaml:"otelagent,omitempty"`
	Otelcol   *Otelcol   `yaml:"otelcol,omitempty"`
}

type Otelagent struct {
	Config map[string]interface{} `yaml:"config,omitempty"`
	Rest   map[string]interface{} `yaml:",inline"`
}

type Otelcol struct {
	Config map[string]interface{} `yaml:"config,omitempty"`
	Rest   map[string]interface{} `yaml:",inline"`
}
