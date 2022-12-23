package tracingconfig

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

func Migrate(input string) (string, error) {
	inputValues, err := parseValues(input)
	if err != nil {
		return "", fmt.Errorf("error parsing input yaml: %v", err)
	}

	if &inputValues.Otelcol != nil {
		outputValues, err := migrate(&inputValues)
		if err != nil {
			return "", fmt.Errorf("error migrating: %v", err)
		}

		buffer := bytes.Buffer{}
		encoder := yaml.NewEncoder(&buffer)
		encoder.SetIndent(2)
		err = encoder.Encode(outputValues)
		fmt.Sprintln(buffer.String())
		fmt.Println("WARNING! Tracing config migrated to v3, please check the output file. For more details see documentation: https://github.com/SumoLogic/sumologic-kubernetes-collection/blob/main/docs/v3-migration-doc.md#tracinginstrumentation-changes")
		return buffer.String(), err
	}

	return input, err
}

func parseValues(input string) (ValuesInput, error) {
	var outputValues ValuesInput
	err := yaml.Unmarshal([]byte(input), &outputValues)
	return outputValues, err
}

func migrate(inputValues *ValuesInput) (ValuesOutput, error) {
	outputValues := ValuesOutput{
		Rest: inputValues.Rest,
	}
	// migrate otelcol source processor to otelcol-instrumentation
	outputValues.OtelcolInstrumentation.Config.Processors.Source = inputValues.Otelcol.Config.Processors.Source
	// migrate otelcol cascading_filter processor to tracesSampler
	outputValues.TracesSampler.Config.Processors.CascadingFilter = inputValues.Otelcol.Config.Processors.CascadingFilter

	return outputValues, nil
}
