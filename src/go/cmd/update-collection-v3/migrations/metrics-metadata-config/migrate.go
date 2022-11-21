package metricsmetadataconfig

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

type InputValues struct {
	Metadata struct {
		Metrics MetricsMetadataInput   `yaml:"metrics,omitempty"`
		Rest    map[string]interface{} `yaml:",inline"`
	} `yaml:"metadata,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}

type MetricsMetadataInput struct {
	Config map[string]interface{} `yaml:"config,omitempty"`
	Rest   map[string]interface{} `yaml:",inline"`
}

type OutputValues struct {
	Metadata struct {
		Metrics MetricsMetadataOutput  `yaml:"metrics,omitempty"`
		Rest    map[string]interface{} `yaml:",inline"`
	} `yaml:"metadata,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}

type MetricsMetadataOutput struct {
	Config struct {
		Merge    map[string]interface{} `yaml:"merge,omitempty"`
		Override map[string]interface{} `yaml:"override,omitempty"`
	} `yaml:"config,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}

func Migrate(inputYaml string) (outputYaml string, err error) {
	inputValues, err := parseValues(inputYaml)
	if err != nil {
		return "", fmt.Errorf("error parsing input yaml: %v", err)
	}

	outputValues, err := migrate(&inputValues)
	if err != nil {
		return "", fmt.Errorf("error migrating: %v", err)
	}

	buffer := bytes.Buffer{}
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	err = encoder.Encode(outputValues)
	return buffer.String(), err
}

func parseValues(inputYaml string) (InputValues, error) {
	var inputValues InputValues
	err := yaml.Unmarshal([]byte(inputYaml), &inputValues)
	return inputValues, err
}

func migrate(inputValues *InputValues) (OutputValues, error) {
	outputValues := OutputValues{
		Rest: inputValues.Rest,
	}
	outputValues.Metadata.Rest = inputValues.Metadata.Rest
	outputValues.Metadata.Metrics.Config.Merge = inputValues.Metadata.Metrics.Config
	return outputValues, nil
}
