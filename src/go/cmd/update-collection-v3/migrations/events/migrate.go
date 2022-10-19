package events

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

func Migrate(yamlV2 string) (yamlV3 string, err error) {
	valuesV2, err := parseValues(yamlV2)
	if err != nil {
		return "", fmt.Errorf("error parsing input yaml: %v", err)
	}

	valuesV3, err := migrate(&valuesV2)
	if err != nil {
		return "", fmt.Errorf("error migrating: %v", err)
	}

	buffer := bytes.Buffer{}
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	err = encoder.Encode(valuesV3)
	return buffer.String(), err
}

func parseValues(yamlV2 string) (ValuesV2, error) {
	var valuesV2 ValuesV2
	err := yaml.Unmarshal([]byte(yamlV2), &valuesV2)
	return valuesV2, err
}

func migrate(valuesV2 *ValuesV2) (ValuesV3, error) {
	valuesV3 := ValuesV3{
		Rest: valuesV2.Rest,
	}
	valuesV3.Sumologic.Rest = valuesV2.Sumologic.Rest
	valuesV3.Fluentd.Rest = valuesV2.Fluentd.Rest
	valuesV3.Fluentd.Persistence = valuesV2.Fluentd.Persistence
	migrateEventsFull(&valuesV3, valuesV2)
	return valuesV3, nil
}
