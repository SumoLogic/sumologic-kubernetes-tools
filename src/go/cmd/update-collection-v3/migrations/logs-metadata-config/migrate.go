package logsmetadataconfig

import (
	"bytes"
	"fmt"

	"github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/helpers"
	"gopkg.in/yaml.v3"
)

type InputValues struct {
	Metadata struct {
		Logs LogsMetadataInput      `yaml:"logs,omitempty"`
		Rest map[string]interface{} `yaml:",inline"`
	} `yaml:"metadata,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}

type LogsMetadataInput struct {
	Config map[string]interface{} `yaml:"config,omitempty"`
	Rest   map[string]interface{} `yaml:",inline"`
}

type OutputValues struct {
	Metadata struct {
		Logs LogsMetadataOutput     `yaml:"logs,omitempty"`
		Rest map[string]interface{} `yaml:",inline"`
	} `yaml:"metadata,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}

type LogsMetadataOutput struct {
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
	_, err = helpers.CheckForConflictsInRest(outputValues)
	if err != nil {
		return "", err
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
	outputValues.Metadata.Logs.Rest = inputValues.Metadata.Logs.Rest
	outputValues.Metadata.Logs.Config.Merge = inputValues.Metadata.Logs.Config
	return outputValues, nil
}
