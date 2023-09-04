package eventsconfigmerge

import (
	"bytes"
	"fmt"

	"github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/helpers"
	"gopkg.in/yaml.v3"
)

type InputValues struct {
	Otelevents struct {
		Config struct {
			Override map[string]interface{} `yaml:"override,omitempty"`
		} `yaml:"config,omitempty"`
		Rest map[string]interface{} `yaml:",inline"`
	} `yaml:"otelevents,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}

type OutputValues struct {
	Otelevents Otelevents             `yaml:"otelevents,omitempty"`
	Rest       map[string]interface{} `yaml:",inline"`
}

type Otelevents struct {
	Config Config                 `yaml:"config,omitempty"`
	Rest   map[string]interface{} `yaml:",inline"`
}

type Config struct {
	Merge    map[string]interface{} `yaml:"merge,omitempty"`
	Override map[string]interface{} `yaml:"override,omitempty"`
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
		Otelevents: Otelevents{
			Config: Config{
				Merge: inputValues.Otelevents.Config.Override,
			},
			Rest: inputValues.Otelevents.Rest,
		},
	}
	return outputValues, nil
}
