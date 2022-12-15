package tracingreplaces

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

var otelcolReplaces []string = []string{
	"processors.sourcea.collector.replace",
	"processors.source.name.replace",
	"processors.source.category.replace",
	"processors.source.category_prefix.replace",
	"processors.source.category_replace_dash.replace",
	"processors.source.exclude_namespace_regex.replace",
	"processors.source.exclude_pod_regex.replace",
	"processors.source.exclude_container_regex.replace",
	"processors.source.exclude_host_regex.replace",
	"processors.resource.cluster.replace",
}

func Migrate(yamlV2 string) (yamlV3 string, err error) {
	valuesV2, err := parseValues(yamlV2)
	if err != nil {
		return "", fmt.Errorf("error parsing input yaml: %v", err)
	}

	if &valuesV2.Otelcol != nil {
		foundOtelcolReplaces := []string{}
		foundOtelcolReplaces, err = findUsedReplaces(valuesV2.Otelcol, otelcolReplaces)
		if err != nil {
			return "", fmt.Errorf("error parsing otelcol configuration: %v", err)
		}
		if len(foundOtelcolReplaces) != 0 {
			fmt.Println("WARNING! Found following special values in otelcol configuration which must be manually migrated:")
			fmt.Println(strings.Join(foundOtelcolReplaces, "\n"))
			fmt.Println("for details please see documentation: https://github.com/SumoLogic/sumologic-kubernetes-collection/blob/main/docs/v3-migration-doc.md#traces")
		}
	}

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

func parseValues(yamlV2 string) (ValuesV2, error) {
	var valuesV2 ValuesV2
	err := yaml.Unmarshal([]byte(yamlV2), &valuesV2)
	return valuesV2, err
}

func parseConfigToString(config Otelcol) (string, error) {
	buffer := bytes.Buffer{}
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	err := encoder.Encode(config)
	if err != nil {
		return "", err
	}
	return buffer.String(), err
}

func findUsedReplaces(config Otelcol, replaces []string) ([]string, error) {
	if &config == nil {
		return []string{}, nil
	}

	confStr, err := parseConfigToString(config)
	if err != nil {
		return []string{}, err
	}

	found := []string{}
	for _, r := range replaces {
		if strings.Contains(confStr, r) {
			found = append(found, fmt.Sprintf(" - %s", r))
		}
	}
	return found, nil
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
