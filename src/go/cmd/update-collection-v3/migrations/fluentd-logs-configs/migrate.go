package fluentdlogsconfigs

import (
	"bytes"
	"fmt"

	"github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/helpers"
	"gopkg.in/yaml.v3"
)

func Migrate(input string) (string, error) {
	valuesInput, err := parseValues(input)
	if err != nil {
		return "", fmt.Errorf("error parsing input yaml: %v", err)
	}

	if valuesInput.Fluentd == nil || valuesInput.Fluentd.Logs == nil {
		// migration of fluentd.logs keys is not needed
		return input, nil
	}

	valuesOutput := migrate(&valuesInput)
	_, err = helpers.CheckForConflictsInRest(valuesOutput)
	if err != nil {
		return "", err
	}

	buffer := bytes.Buffer{}
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	err = encoder.Encode(valuesOutput)
	return buffer.String(), err
}

func parseValues(input string) (ValuesInput, error) {
	var valuesInput ValuesInput
	err := yaml.Unmarshal([]byte(input), &valuesInput)
	return valuesInput, err
}

func migrate(valuesInput *ValuesInput) ValuesOutput {
	return ValuesOutput{
		Rest:      valuesInput.Rest,
		Sumologic: createSumologic(valuesInput),
		Fluentd:   createFluentdLogs(valuesInput),
	}
}
