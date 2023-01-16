package fluentdautoscaling

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

func Migrate(input string) (string, error) {
	valuesInput, err := parseValues(input)
	if err != nil {
		return "", fmt.Errorf("error parsing input yaml: %v", err)
	}

	valuesOutput := migrate(&valuesInput)

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
	valuesOutput := ValuesOutput{
		Rest:     valuesInput.Rest,
		Fluentd:  valuesInput.Fluentd,
		Metadata: valuesInput.Metadata,
	}

	// we only do something if autoscaling is set for either FluentD logs or metrics
	if valuesInput.Fluentd == nil {
		return valuesOutput
	}

	migrateMetricsAutoscaling(valuesInput, &valuesOutput)
	migrateLogsAutoscaling(valuesInput, &valuesOutput)

	return valuesOutput
}

func migrateMetricsAutoscaling(valuesInput *ValuesInput, valuesOutput *ValuesOutput) {
	if valuesInput.Fluentd.Metrics == nil ||
		valuesInput.Fluentd.Metrics.Autoscaling == nil ||
		valuesInput.Fluentd.Metrics.Autoscaling.Enabled == nil ||
		!*valuesInput.Fluentd.Metrics.Autoscaling.Enabled {
		return
	}

	if valuesOutput.Metadata == nil {
		valuesOutput.Metadata = &Metadata{}
	}

	if valuesOutput.Metadata.Metrics == nil {
		valuesOutput.Metadata.Metrics = &MetadataMetrics{}
	}

	if valuesOutput.Metadata.Metrics.Autoscaling == nil {
		valuesOutput.Metadata.Metrics.Autoscaling = &Autoscaling{}
	}

	if valuesOutput.Metadata.Metrics.Autoscaling.Enabled == nil {
		valuesOutput.Metadata.Metrics.Autoscaling.Enabled = new(bool)
		*valuesOutput.Metadata.Metrics.Autoscaling.Enabled = true
	}
}

func migrateLogsAutoscaling(valuesInput *ValuesInput, valuesOutput *ValuesOutput) {
	if valuesInput.Fluentd.Logs == nil ||
		valuesInput.Fluentd.Logs.Autoscaling == nil ||
		valuesInput.Fluentd.Logs.Autoscaling.Enabled == nil ||
		!*valuesInput.Fluentd.Logs.Autoscaling.Enabled {
		return
	}

	if valuesOutput.Metadata == nil {
		valuesOutput.Metadata = &Metadata{}
	}

	if valuesOutput.Metadata.Logs == nil {
		valuesOutput.Metadata.Logs = &MetadataLogs{}
	}

	if valuesOutput.Metadata.Logs.Autoscaling == nil {
		valuesOutput.Metadata.Logs.Autoscaling = &Autoscaling{}
	}

	if valuesOutput.Metadata.Logs.Autoscaling.Enabled == nil {
		valuesOutput.Metadata.Logs.Autoscaling.Enabled = new(bool)
		*valuesOutput.Metadata.Logs.Autoscaling.Enabled = true
	}
}
