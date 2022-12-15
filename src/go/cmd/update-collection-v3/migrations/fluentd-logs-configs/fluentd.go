package fluentdlogsconfigs

func createSumologic(valuesV2 *ValuesV2) *SumologicV3 {
	if valuesV2.Sumologic == nil && valuesV2.Fluentd == nil {
		return nil
	}

	sumoLogicV3 := &SumologicV3{}
	if valuesV2.Fluentd != nil {
		sumoLogicV3.Logs = createSumologicLogs(valuesV2.Fluentd.Logs, valuesV2.Sumologic.Logs)
	}

	if valuesV2.Sumologic != nil {
		sumoLogicV3.Rest = valuesV2.Sumologic.Rest
	}
	return sumoLogicV3
}

func createSumologicLogs(fluentdLogsV2 *FluentdLogs, sumologicLogsV2 *SumologicLogsV2) *SumologicLogsV3 {
	sumologicLogsV3 := &SumologicLogsV3{}

	if fluentdLogsV2 != nil {
		sumologicLogsV3.Container = createSumologicLogsContainer(fluentdLogsV2.Containers, sumologicLogsV2.Container)
		sumologicLogsV3.Systemd = createSumologicLogsConfig(fluentdLogsV2.Systemd, sumologicLogsV2.Systemd)
		sumologicLogsV3.Kubelet = createSumologicLogsConfig(fluentdLogsV2.Kubelet, nil) // set nil because in v2 there was not any configuration under sumologic.logs.kubelet
		sumologicLogsV3.Default = createSumologicLogsConfig(fluentdLogsV2.Default, nil) // set nil because in v2 there was not any configuration under sumologic.logs.defaultFluentd
	}

	if sumologicLogsV2 != nil {
		sumologicLogsV3.Rest = sumologicLogsV2.Rest
	}
	return sumologicLogsV3
}

func isLogsMigrationNeeded(configV2 *LogsConfig) bool {
	if configV2 == nil {
		return false
	} else if configV2.SourceName != nil ||
		configV2.SourceCategory != nil ||
		configV2.SourceCategoryPrefix != nil ||
		configV2.SourceCategoryReplaceDash != nil ||
		configV2.ExcludeFacilityRegex != nil ||
		configV2.ExcludeHostRegex != nil ||
		configV2.ExcludePriorityRegex != nil ||
		configV2.ExcludeUnitRegex != nil {
		return true
	}
	return false
}

func createSumologicLogsConfig(fluentdLogsConfigV2 *LogsConfig, sumologicLogsConfigV2 map[string]interface{}) *LogsConfig {
	if !isLogsMigrationNeeded(fluentdLogsConfigV2) {
		if sumologicLogsConfigV2 != nil {
			return &LogsConfig{Rest: sumologicLogsConfigV2}
		}
		return nil
	}

	return &LogsConfig{
		SourceName:                fluentdLogsConfigV2.SourceName,
		SourceCategory:            fluentdLogsConfigV2.SourceCategory,
		SourceCategoryPrefix:      fluentdLogsConfigV2.SourceCategoryPrefix,
		SourceCategoryReplaceDash: fluentdLogsConfigV2.SourceCategoryReplaceDash,
		ExcludeFacilityRegex:      fluentdLogsConfigV2.ExcludeFacilityRegex,
		ExcludeHostRegex:          fluentdLogsConfigV2.ExcludeHostRegex,
		ExcludePriorityRegex:      fluentdLogsConfigV2.ExcludePriorityRegex,
		ExcludeUnitRegex:          fluentdLogsConfigV2.ExcludeUnitRegex,
		Rest:                      sumologicLogsConfigV2,
	}
}

func isContainerLogsMigrationNeeded(configV2 *ContainersLogsConfig) bool {
	if configV2 == nil {
		return false
	} else if configV2.SourceName != nil ||
		configV2.SourceCategory != nil ||
		configV2.SourceCategoryPrefix != nil ||
		configV2.SourceCategoryReplaceDash != nil ||
		configV2.ExcludeFacilityRegex != nil ||
		configV2.ExcludeHostRegex != nil ||
		configV2.ExcludePriorityRegex != nil ||
		configV2.ExcludeUnitRegex != nil ||
		configV2.PerContainerAnnotationsEnabled != nil ||
		configV2.PerContainerAnnotationPrefixes != nil {
		return true
	}
	return false
}

func createSumologicLogsContainer(fluendLogsContainersV2 *ContainersLogsConfig, sumologicLogsContainerV2 map[string]interface{}) *ContainersLogsConfig {
	if !isContainerLogsMigrationNeeded(fluendLogsContainersV2) {
		if sumologicLogsContainerV2 != nil {
			return &ContainersLogsConfig{Rest: sumologicLogsContainerV2}
		}
		return nil
	}

	return &ContainersLogsConfig{
		SourceName:                     fluendLogsContainersV2.SourceName,
		SourceCategory:                 fluendLogsContainersV2.SourceCategory,
		SourceCategoryPrefix:           fluendLogsContainersV2.SourceCategoryPrefix,
		SourceCategoryReplaceDash:      fluendLogsContainersV2.SourceCategoryReplaceDash,
		ExcludeFacilityRegex:           fluendLogsContainersV2.ExcludeFacilityRegex,
		ExcludeHostRegex:               fluendLogsContainersV2.ExcludeHostRegex,
		ExcludePriorityRegex:           fluendLogsContainersV2.ExcludePriorityRegex,
		ExcludeUnitRegex:               fluendLogsContainersV2.ExcludeUnitRegex,
		PerContainerAnnotationsEnabled: fluendLogsContainersV2.PerContainerAnnotationsEnabled,
		PerContainerAnnotationPrefixes: fluendLogsContainersV2.PerContainerAnnotationPrefixes,
		Rest:                           sumologicLogsContainerV2,
	}
}

func isLogRestEmpty(config *LogsConfig) bool {
	if config == nil || config.Rest == nil {
		return true
	}
	return false
}

func isContainersLogRestEmpty(config *ContainersLogsConfig) bool {
	if config == nil || config.Rest == nil {
		return true
	}
	return false
}

func isFluentdV3Empty(fluentdV2 *Fluentd) bool {
	if fluentdV2 == nil ||
		(fluentdV2.Rest == nil &&
			isContainersLogRestEmpty(fluentdV2.Logs.Containers) &&
			isLogRestEmpty(fluentdV2.Logs.Systemd) &&
			isLogRestEmpty(fluentdV2.Logs.Kubelet) &&
			isLogRestEmpty(fluentdV2.Logs.Default)) {
		return true
	}
	return false
}

func createFluentdLogsContainersConfig(containersConfigV2 *ContainersLogsConfig) *ContainersLogsConfig {
	if containersConfigV2 != nil && containersConfigV2.Rest != nil {
		return &ContainersLogsConfig{
			Rest: containersConfigV2.Rest,
		}
	}
	return nil
}

func createFluentdLogsConfig(logsConfigV2 *LogsConfig) *LogsConfig {
	if logsConfigV2 != nil && logsConfigV2.Rest != nil {
		return &LogsConfig{
			Rest: logsConfigV2.Rest,
		}
	}
	return nil
}

func createFluentdLogs(valuesV2 *ValuesV2) *Fluentd {
	if isFluentdV3Empty(valuesV2.Fluentd) {
		return nil
	}

	if valuesV2.Fluentd.Logs == nil {
		return &Fluentd{Rest: valuesV2.Fluentd.Rest}
	}

	return &Fluentd{
		Rest: valuesV2.Fluentd.Rest,
		Logs: &FluentdLogs{
			Containers: createFluentdLogsContainersConfig(valuesV2.Fluentd.Logs.Containers),
			Systemd:    createFluentdLogsConfig(valuesV2.Fluentd.Logs.Systemd),
			Kubelet:    createFluentdLogsConfig(valuesV2.Fluentd.Logs.Kubelet),
			Default:    createFluentdLogsConfig(valuesV2.Fluentd.Logs.Default),
		},
	}
}
