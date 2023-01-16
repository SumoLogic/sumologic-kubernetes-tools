package fluentdautoscaling

type ValuesInput struct {
	Metadata *Metadata              `yaml:"metadata,omitempty"`
	Fluentd  *Fluentd               `yaml:"fluentd,omitempty"`
	Rest     map[string]interface{} `yaml:",inline"`
}
