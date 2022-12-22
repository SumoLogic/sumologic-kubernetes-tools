package tracingobjectchanges

type ValuesV2 struct {
	Sumologic struct {
		Traces struct {
			Enabled bool `yaml:"enabled,omitempty"`
		} `yaml:"traces,omitempty"`
	} `yaml:"sumologic,omitempty"`
	Otelcol   map[string]interface{} `yaml:"otelcol,omitempty"`
	Otelagent struct {
		Enabled bool `yaml:"enabled,omitempty"`
	} `yaml:"otelagent,omitempty"`
	Otelgateway struct {
		Enabled bool `yaml:"enabled,omitempty"`
	} `yaml:"otelgateway,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}
