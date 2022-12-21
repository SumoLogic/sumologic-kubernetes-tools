package tracingconfig

import (
	"bytes"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

func Migrate(yamlV2 string) (yamlV3 string, err error) {
	valuesV2, err := parseValues(yamlV2)
	if err != nil {
		return "", fmt.Errorf("error parsing input yaml: %v", err)
	}
	fmt.Println(reflect.TypeOf(valuesV2))

	if &valuesV2.Otelcol == nil {
		valuesV3, err := migrate(&valuesV2)
		if err != nil {
			return "", fmt.Errorf("error migrating: %v", err)
		}

		buffer := bytes.Buffer{}
		encoder := yaml.NewEncoder(&buffer)
		encoder.SetIndent(2)
		err = encoder.Encode(valuesV3)
		fmt.Sprintln(buffer.String())
		return buffer.String(), err
	}

	return yamlV2, err
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
	// migrate otelcol source processor to otelcol-instrumentation
	valuesV3.OtelcolInstrumentation.Config.Processors.Source = valuesV2.Otelcol.Config.Processors.Source
	// migrate otelcol cascading_filter processor to tracesSampler
	valuesV3.TracesSampler.Config.Processors.CascadingFilter = valuesV2.Otelcol.Config.Processors.CascadingFilter
	return valuesV3, nil
}
