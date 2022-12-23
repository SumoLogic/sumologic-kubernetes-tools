package tracingobjectchanges

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func Migrate(input string) (string, error) {
	inputValues, err := parseValues(input)
	if err != nil {
		return "", fmt.Errorf("error parsing input yaml: %v", err)
	}

	if &inputValues.Sumologic.Traces != nil {
		if inputValues.Sumologic.Traces.Enabled == true {
			fmt.Println("WARNING! Found enabled otelcol, for details please see documentation: https://github.com/SumoLogic/sumologic-kubernetes-collection/blob/main/docs/v3-migration-doc.md#tracinginstrumentation-changes")
		}

	}

	if &inputValues.Otelagent != nil {
		if inputValues.Otelagent.Enabled == true {
			fmt.Println("WARNING! Found enabled otelagent, for details please see documentation: https://github.com/SumoLogic/sumologic-kubernetes-collection/blob/main/docs/v3-migration-doc.md#tracinginstrumentation-changes")
		}
	}

	if &inputValues.Otelgateway != nil {
		if inputValues.Otelgateway.Enabled == true {
			fmt.Println("WARNING! Found enabled otelgateway, for details please see documentation: https://github.com/SumoLogic/sumologic-kubernetes-collection/blob/main/docs/v3-migration-doc.md#tracinginstrumentation-changes")
		}
	}

	return input, err
}

func parseValues(input string) (ValuesInput, error) {
	var inputValues ValuesInput
	err := yaml.Unmarshal([]byte(input), &inputValues)
	return inputValues, err
}
