package fluentdlogsconfigs

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

	if valuesV2.Fluentd == nil || valuesV2.Fluentd.Logs == nil {
		// migration of fluentd.logs keys is not needed
		return yamlV2, nil
	}

	valuesV3 := migrate(&valuesV2)

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

func migrate(valuesV2 *ValuesV2) ValuesV3 {
	return ValuesV3{
		Rest:      valuesV2.Rest,
		Sumologic: setSumologic(valuesV2),
		Fluentd:   setFluentdLogs(valuesV2),
	}
}
