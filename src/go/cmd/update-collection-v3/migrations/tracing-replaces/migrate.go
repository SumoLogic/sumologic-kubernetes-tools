package tracingreplaces

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/helpers"
	"gopkg.in/yaml.v3"
)

var otelcolReplaces []string = []string{
	"processors.source.collector.replace",
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

func Migrate(input string) (string, error) {
	inputValues, err := parseValues(input)
	if err != nil {
		return "", fmt.Errorf("error parsing input yaml: %v", err)
	}

	if &inputValues.Otelcol != nil {
		foundOtelcolReplaces := []string{}
		foundOtelcolReplaces, err = findUsedReplaces(inputValues.Otelcol, otelcolReplaces)
		if err != nil {
			return "", fmt.Errorf("error parsing otelcol configuration: %v", err)
		}
		if len(foundOtelcolReplaces) != 0 {
			fmt.Println("WARNING! Found following special values in otelcol configuration which must be manually migrated:")
			fmt.Println(strings.Join(foundOtelcolReplaces, "\n"))
			fmt.Println("for details please see documentation: https://github.com/SumoLogic/sumologic-kubernetes-collection/blob/main/docs/v3-migration-doc.md#replace-special-configuration-values-marked-by-replace-suffix")
		}
	}

	return input, err
}

func parseValues(input string) (ValuesInput, error) {
	var inputValues ValuesInput
	err := yaml.Unmarshal([]byte(input), &inputValues)
	return inputValues, err
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
	_, err := helpers.CheckForConflictsInRest(config)
	if err != nil {
		return []string{}, err
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
