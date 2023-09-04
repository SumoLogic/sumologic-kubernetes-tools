package kubestatemetricscollectors

import (
	"bytes"
	"fmt"

	"github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/helpers"
	"gopkg.in/yaml.v3"
)

func Migrate(yamlV2 string) (yamlV3 string, err error) {
	valuesV2, err := parseValues(yamlV2)
	if err != nil {
		return "", fmt.Errorf("error parsing input yaml: %v", err)
	}

	valuesV3, err := migrate(&valuesV2)
	if err != nil {
		return "", fmt.Errorf("error migrating: %v", err)
	}
	_, err = helpers.CheckForConflictsInRest(valuesV3)
	if err != nil {
		return "", err
	}

	buffer := bytes.Buffer{}
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	err = encoder.Encode(valuesV3)
	return buffer.String(), err
}

func parseValues(yamlV2 string) (ValuesV2, error) {
	var valuesV2 ValuesV2
	err := yaml.Unmarshal([]byte(yamlV2), &valuesV2)
	return valuesV2, err
}

func migrate(valuesV2 *ValuesV2) (ValuesV3, error) {
	valuesV3 := ValuesV3{
		Rest:                valuesV2.Rest,
		KubePrometheusStack: migrateKubePrometheusStack(valuesV2.KubePrometheusStack),
	}
	return valuesV3, nil
}

func migrateKubePrometheusStack(valuesV2 *KubePrometheusStackV2) *KubePrometheusStackV3 {
	if valuesV2 == nil {
		return nil
	}

	return &KubePrometheusStackV3{
		Rest:             valuesV2.Rest,
		KubeStateMetrics: migrateKubeStateMetrics(valuesV2.KubeStateMetrics),
	}
}

func migrateKubeStateMetrics(valuesV2 *KubeStateMetricsV2) *KubeStateMetricsV3 {
	if valuesV2 == nil {
		return nil
	}

	return &KubeStateMetricsV3{
		Rest:       valuesV2.Rest,
		Collectors: migrateKubeStateMetricsCollectors(valuesV2.Collectors),
	}
}

func migrateKubeStateMetricsCollectors(collectors *map[string]bool) *[]string {
	if collectors == nil {
		return nil
	}

	kubeStateMetricsCollectorsList := []string{
		"certificatesigningrequests",
		"configmaps",
		"cronjobs",
		"daemonsets",
		"deployments",
		"endpoints",
		"horizontalpodautoscalers",
		"ingresses",
		"jobs",
		"limitranges",
		"mutatingwebhookconfigurations",
		"namespaces",
		"networkpolicies",
		"nodes",
		"persistentvolumeclaims",
		"persistentvolumes",
		"poddisruptionbudgets",
		"pods",
		"replicasets",
		"replicationcontrollers",
		"resourcequotas",
		"secrets",
		"services",
		"statefulsets",
		"storageclasses",
		"validatingwebhookconfigurations",
		"volumeattachments",
	}

	returnList := []string{}
	disabled := false

	for _, key := range kubeStateMetricsCollectorsList {
		if value, ok := (*collectors)[key]; ok && !value {
			disabled = true
			continue
		}
		returnList = append(returnList, key)
	}

	if !disabled {
		return nil
	}

	return &returnList
}
