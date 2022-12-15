package fluentdlogsconfigs

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

func Migrate(input string) (string, error) {
	valuesV2, err := parseValues(input)
	if err != nil {
		return "", fmt.Errorf("error parsing input yaml: %v", err)
	}

	if valuesV2.Fluentd == nil || valuesV2.Fluentd.Logs == nil {
		// migration of fluentd.logs keys is not needed
		return input, nil
	}

	valuesV3 := migrate(&valuesV2)

	buffer := bytes.Buffer{}
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	err = encoder.Encode(valuesV3)
	return buffer.String(), err
}

func parseValues(input string) (ValuesV2, error) {
	var valuesV2 ValuesV2
	err := yaml.Unmarshal([]byte(input), &valuesV2)
	return valuesV2, err
}

func migrate(valuesV2 *ValuesV2) ValuesV3 {
	return ValuesV3{
		Rest:      valuesV2.Rest,
		Sumologic: createSumologic(valuesV2),
		Fluentd:   createFluentdLogs(valuesV2),
	}
}
