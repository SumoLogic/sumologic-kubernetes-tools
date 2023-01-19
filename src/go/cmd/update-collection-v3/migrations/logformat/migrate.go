package logformat

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

func Migrate(inputYaml string) (outputYaml string, err error) {
	inputValues, err := parseValues(inputYaml)
	if err != nil {
		return "", fmt.Errorf("error parsing input yaml: %v", err)
	}

	outputValues := migrate(&inputValues)
	if err != nil {
		return "", fmt.Errorf("error migrating: %v", err)
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

func migrate(inputValues *InputValues) OutputValues {
	outputValues := *inputValues

	if outputValues.Fluentd == nil ||
		outputValues.Fluentd.Logs == nil ||
		outputValues.Fluentd.Logs.Output == nil ||
		outputValues.Fluentd.Logs.Output.LogFormat == nil {
		// not set for Fluentd, leave things as is
		return outputValues
	}

	logFormat := *outputValues.Fluentd.Logs.Output.LogFormat

	if outputValues.Sumologic == nil {
		outputValues.Sumologic = &Sumologic{}
	}

	if outputValues.Sumologic.Logs == nil {
		outputValues.Sumologic.Logs = &SumologicLogs{}
	}

	if outputValues.Sumologic.Logs.Container == nil {
		outputValues.Sumologic.Logs.Container = &SumologicLogsContainer{}
	}

	if outputValues.Sumologic.Logs.Container.Format == nil {
		outputValues.Sumologic.Logs.Container.Format = new(string)
		*outputValues.Sumologic.Logs.Container.Format = logFormat
	}

	// remove log format for Fluentd and clean up empty structs
	outputValues.Fluentd.Logs.Output.LogFormat = nil
	if len(outputValues.Fluentd.Logs.Output.Rest) == 0 {
		outputValues.Fluentd.Logs.Output = nil
	}
	if len(outputValues.Fluentd.Logs.Rest) == 0 {
		outputValues.Fluentd.Logs = nil
	}
	if len(outputValues.Fluentd.Rest) == 0 {
		outputValues.Fluentd = nil
	}

	return outputValues
}
