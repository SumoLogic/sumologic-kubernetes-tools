package disablethanos

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

func Migrate(inputYaml string) (string, error) {
	var inputValues InputValues
	err := parseYaml(inputYaml, &inputValues)
	if err != nil {
		return "", err
	}
	outputValues, err := migrate(&inputValues)
	if err != nil {
		return "", err
	}
	return formatYaml(outputValues)
}

func parseYaml[T any](yamlString string, structure T) error {
	err := yaml.Unmarshal([]byte(yamlString), &structure)
	return err
}

func formatYaml[T any](inputStructure *T) (string, error) {
	buffer := bytes.Buffer{}
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	err := encoder.Encode(inputStructure)
	return buffer.String(), err
}

func migrate(input *InputValues) (*OutputValues, error) {
	output := &OutputValues{
		KubePrometheusStack: &KubePrometheusStackOutput{
			Prometheus: &PrometheusOutput{
				PrometheusSpec: &PrometheusSpecOutput{},
			},
		},
	}

	if input.KubePrometheusStack != nil {
		if input.KubePrometheusStack.Prometheus != nil {
			if input.KubePrometheusStack.Prometheus.PrometheusSpec != nil {
				output.KubePrometheusStack.Prometheus.PrometheusSpec.Rest = input.KubePrometheusStack.Prometheus.PrometheusSpec.Rest
			}
			output.KubePrometheusStack.Prometheus.Rest = input.KubePrometheusStack.Prometheus.Rest
		}
		output.KubePrometheusStack.Rest = input.KubePrometheusStack.Rest
	}
	output.Rest = input.Rest

	if output.KubePrometheusStack.Prometheus.PrometheusSpec.Rest == nil {
		output.KubePrometheusStack.Prometheus.PrometheusSpec = nil
	}
	if output.KubePrometheusStack.Prometheus.PrometheusSpec == nil && output.KubePrometheusStack.Prometheus.Rest == nil {
		output.KubePrometheusStack.Prometheus = nil
	}
	if output.KubePrometheusStack.Prometheus == nil && output.KubePrometheusStack.Rest == nil {
		output.KubePrometheusStack = nil
	}

	return output, nil
}

type InputValues struct {
	KubePrometheusStack *KubePrometheusStackInput `yaml:"kube-prometheus-stack,omitempty"`
	Rest                map[string]interface{}    `yaml:",inline,omitempty"`
}

type KubePrometheusStackInput struct {
	Prometheus *PrometheusInput       `yaml:",omitempty"`
	Rest       map[string]interface{} `yaml:",inline,omitempty"`
}

type PrometheusInput struct {
	PrometheusSpec *PrometheusSpecInput   `yaml:"prometheusSpec,omitempty"`
	Rest           map[string]interface{} `yaml:",inline,omitempty"`
}

type PrometheusSpecInput struct {
	Thanos map[string]interface{} `yaml:",omitempty"`
	Rest   map[string]interface{} `yaml:",inline,omitempty"`
}

type OutputValues struct {
	KubePrometheusStack *KubePrometheusStackOutput `yaml:"kube-prometheus-stack,omitempty"`
	Rest                map[string]interface{}     `yaml:",inline,omitempty"`
}

type KubePrometheusStackOutput struct {
	Prometheus *PrometheusOutput      `yaml:",omitempty"`
	Rest       map[string]interface{} `yaml:",inline,omitempty"`
}

type PrometheusOutput struct {
	PrometheusSpec *PrometheusSpecOutput  `yaml:"prometheusSpec,omitempty"`
	Rest           map[string]interface{} `yaml:",inline,omitempty"`
}

type PrometheusSpecOutput struct {
	Thanos map[string]interface{} `yaml:",omitempty"`
	Rest   map[string]interface{} `yaml:",inline,omitempty"`
}
