package main

// TODO: implement
func migrate(valuesV2 *ValuesV2) (ValuesV3, error) {
	valuesV3 := ValuesV3{
		Rest:                valuesV2.Rest,
		KubePrometheusStack: migrateKubePrometheusStack(valuesV2.KubePrometheusStack),
	}
	valuesV3.Sumologic.Rest = valuesV2.Sumologic.Rest
	valuesV3.Fluentd.Rest = valuesV2.Fluentd.Rest
	valuesV3.Fluentd.Persistence = valuesV2.Fluentd.Persistence
	migrateEventsFull(&valuesV3, valuesV2)
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

	for _, key := range kubeStateMetricsCollectorsList {
		if value, ok := (*collectors)[key]; ok && !value {
			continue
		}
		returnList = append(returnList, key)
	}
	return &returnList
}
