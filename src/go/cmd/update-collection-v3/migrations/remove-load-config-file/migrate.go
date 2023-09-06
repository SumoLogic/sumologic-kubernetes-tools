package removeloadconfigfile

import (
	"bytes"
	"fmt"

	"github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/helpers"
	"gopkg.in/yaml.v3"
)

type Values struct {
	Sumologic struct {
		Cluster Cluster                `yaml:"cluster,omitempty"`
		Rest    map[string]interface{} `yaml:",inline"`
	} `yaml:"sumologic,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}

type Cluster struct {
	LoadConfigFile *bool                  `yaml:"load_config_file,omitempty"`
	Rest           map[string]interface{} `yaml:",inline"`
}

func Migrate(inputYaml string) (outputYaml string, err error) {
	values, err := parseValues(inputYaml)
	if err != nil {
		return "", fmt.Errorf("error parsing input yaml: %v", err)
	}

	outputValues, err := migrate(&values)
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

func parseValues(inputYaml string) (Values, error) {
	var inputValues Values
	err := yaml.Unmarshal([]byte(inputYaml), &inputValues)
	return inputValues, err
}

func migrate(values *Values) (Values, error) {
	values.Sumologic.Cluster.LoadConfigFile = nil
	return *values, nil
}
