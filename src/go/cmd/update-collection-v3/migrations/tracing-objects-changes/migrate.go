package tracingobjectchanges

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func Migrate(yamlV2 string) (yamlV3 string, err error) {
	valuesV2, err := parseValues(yamlV2)
	if err != nil {
		return "", fmt.Errorf("error parsing input yaml: %v", err)
	}

	if &valuesV2.Sumologic.Traces != nil {
		if valuesV2.Sumologic.Traces.Enabled == true {
			fmt.Println("WARNING! Found enabled otelcol, for details please see documentation: https://github.com/SumoLogic/sumologic-kubernetes-collection/blob/main/docs/v3-migration-doc.md#tracinginstrumentation-changes")
		}

	}

	if &valuesV2.Otelagent != nil {
		if valuesV2.Otelagent.Enabled == true {
			fmt.Println("WARNING! Found enabled otelagent, for details please see documentation: https://github.com/SumoLogic/sumologic-kubernetes-collection/blob/main/docs/v3-migration-doc.md#tracinginstrumentation-changes")
		}
	}

	if &valuesV2.Otelgateway != nil {
		if valuesV2.Otelgateway.Enabled == true {
			fmt.Println("WARNING! Found enabled otelgateway, for details please see documentation: https://github.com/SumoLogic/sumologic-kubernetes-collection/blob/main/docs/v3-migration-doc.md#tracinginstrumentation-changes")
		}
	}

	return yamlV2, err
}

func parseValues(yamlV2 string) (ValuesV2, error) {
	var valuesV2 ValuesV2
	err := yaml.Unmarshal([]byte(yamlV2), &valuesV2)
	return valuesV2, err
}
