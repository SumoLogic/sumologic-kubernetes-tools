package fluentdlogsconfigs

func createSumologic(valuesInput *ValuesInput) *SumologicOutput {
	if valuesInput.Sumologic == nil && valuesInput.Fluentd == nil {
		return nil
	}

	sumoLogicOutput := &SumologicOutput{}
	if valuesInput.Fluentd != nil {
		sumoLogicOutput.Logs = createSumologicLogs(valuesInput.Fluentd.Logs, valuesInput.Sumologic.Logs)
	}

	if valuesInput.Sumologic != nil {
		sumoLogicOutput.Rest = valuesInput.Sumologic.Rest
	}
	return sumoLogicOutput
}

func createSumologicLogs(fluentdLogsInput *FluentdLogs, sumologicLogsInput *SumologicLogsInput) *SumologicLogsOutput {
	sumologicLogsOutput := &SumologicLogsOutput{}

	if fluentdLogsInput != nil {
		var sumologicLogsInputContainer map[string]interface{}
		var sumologicLogsInputSystemd map[string]interface{}
		if sumologicLogsInput != nil {
			sumologicLogsInputContainer = sumologicLogsInput.Container
			sumologicLogsInputSystemd = sumologicLogsInput.Systemd
		}
		sumologicLogsOutput.Container = createSumologicLogsContainer(fluentdLogsInput.Containers, sumologicLogsInputContainer)
		sumologicLogsOutput.Systemd = createSumologicLogsConfig(fluentdLogsInput.Systemd, sumologicLogsInputSystemd)
		sumologicLogsOutput.Kubelet = createSumologicLogsConfig(fluentdLogsInput.Kubelet, nil) // set nil because in v2 there was not any configuration under sumologic.logs.kubelet
		sumologicLogsOutput.Default = createSumologicLogsConfig(fluentdLogsInput.Default, nil) // set nil because in v2 there was not any configuration under sumologic.logs.defaultFluentd
	}

	if sumologicLogsInput != nil {
		sumologicLogsOutput.Rest = sumologicLogsInput.Rest
	}
	return sumologicLogsOutput
}

func isLogsMigrationNeeded(configInput *LogsConfig) bool {
	if configInput == nil {
		return false
	} else if configInput.SourceName != nil ||
		configInput.SourceCategory != nil ||
		configInput.SourceCategoryPrefix != nil ||
		configInput.SourceCategoryReplaceDash != nil ||
		configInput.ExcludeFacilityRegex != nil ||
		configInput.ExcludeHostRegex != nil ||
		configInput.ExcludePriorityRegex != nil ||
		configInput.ExcludeUnitRegex != nil {
		return true
	}
	return false
}

func createSumologicLogsConfig(fluentdLogsConfigInput *LogsConfig, sumologicLogsConfigInput map[string]interface{}) *LogsConfig {
	if !isLogsMigrationNeeded(fluentdLogsConfigInput) {
		if sumologicLogsConfigInput != nil {
			return &LogsConfig{Rest: sumologicLogsConfigInput}
		}
		return nil
	}

	return &LogsConfig{
		SourceName:                fluentdLogsConfigInput.SourceName,
		SourceCategory:            fluentdLogsConfigInput.SourceCategory,
		SourceCategoryPrefix:      fluentdLogsConfigInput.SourceCategoryPrefix,
		SourceCategoryReplaceDash: fluentdLogsConfigInput.SourceCategoryReplaceDash,
		ExcludeFacilityRegex:      fluentdLogsConfigInput.ExcludeFacilityRegex,
		ExcludeHostRegex:          fluentdLogsConfigInput.ExcludeHostRegex,
		ExcludePriorityRegex:      fluentdLogsConfigInput.ExcludePriorityRegex,
		ExcludeUnitRegex:          fluentdLogsConfigInput.ExcludeUnitRegex,
		Rest:                      sumologicLogsConfigInput,
	}
}

func isContainerLogsMigrationNeeded(configInput *ContainersLogsConfig) bool {
	if configInput == nil {
		return false
	} else if configInput.SourceName != nil ||
		configInput.SourceCategory != nil ||
		configInput.SourceCategoryPrefix != nil ||
		configInput.SourceCategoryReplaceDash != nil ||
		configInput.ExcludeFacilityRegex != nil ||
		configInput.ExcludeHostRegex != nil ||
		configInput.ExcludePriorityRegex != nil ||
		configInput.ExcludeUnitRegex != nil ||
		configInput.PerContainerAnnotationsEnabled != nil ||
		configInput.PerContainerAnnotationPrefixes != nil {
		return true
	}
	return false
}

func createSumologicLogsContainer(fluendLogsContainersInput *ContainersLogsConfig, sumologicLogsContainerInput map[string]interface{}) *ContainersLogsConfig {
	if !isContainerLogsMigrationNeeded(fluendLogsContainersInput) {
		if sumologicLogsContainerInput != nil {
			return &ContainersLogsConfig{Rest: sumologicLogsContainerInput}
		}
		return nil
	}

	return &ContainersLogsConfig{
		SourceName:                     fluendLogsContainersInput.SourceName,
		SourceCategory:                 fluendLogsContainersInput.SourceCategory,
		SourceCategoryPrefix:           fluendLogsContainersInput.SourceCategoryPrefix,
		SourceCategoryReplaceDash:      fluendLogsContainersInput.SourceCategoryReplaceDash,
		ExcludeFacilityRegex:           fluendLogsContainersInput.ExcludeFacilityRegex,
		ExcludeHostRegex:               fluendLogsContainersInput.ExcludeHostRegex,
		ExcludePriorityRegex:           fluendLogsContainersInput.ExcludePriorityRegex,
		ExcludeUnitRegex:               fluendLogsContainersInput.ExcludeUnitRegex,
		PerContainerAnnotationsEnabled: fluendLogsContainersInput.PerContainerAnnotationsEnabled,
		PerContainerAnnotationPrefixes: fluendLogsContainersInput.PerContainerAnnotationPrefixes,
		Rest:                           sumologicLogsContainerInput,
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

func isFluentdOutputEmpty(fluentdInput *Fluentd) bool {
	if fluentdInput == nil ||
		(fluentdInput.Rest == nil &&
			isContainersLogRestEmpty(fluentdInput.Logs.Containers) &&
			isLogRestEmpty(fluentdInput.Logs.Systemd) &&
			isLogRestEmpty(fluentdInput.Logs.Kubelet) &&
			isLogRestEmpty(fluentdInput.Logs.Default)) {
		return true
	}
	return false
}

func createFluentdLogsContainersConfig(containersConfigInput *ContainersLogsConfig) *ContainersLogsConfig {
	if containersConfigInput != nil && containersConfigInput.Rest != nil {
		return &ContainersLogsConfig{
			Rest: containersConfigInput.Rest,
		}
	}
	return nil
}

func createFluentdLogsConfig(logsConfigInput *LogsConfig) *LogsConfig {
	if logsConfigInput != nil && logsConfigInput.Rest != nil {
		return &LogsConfig{
			Rest: logsConfigInput.Rest,
		}
	}
	return nil
}

func createFluentdLogs(valuesInput *ValuesInput) *Fluentd {
	if isFluentdOutputEmpty(valuesInput.Fluentd) {
		return nil
	}

	if valuesInput.Fluentd.Logs == nil {
		return &Fluentd{Rest: valuesInput.Fluentd.Rest}
	}

	return &Fluentd{
		Rest: valuesInput.Fluentd.Rest,
		Logs: &FluentdLogs{
			Containers: createFluentdLogsContainersConfig(valuesInput.Fluentd.Logs.Containers),
			Systemd:    createFluentdLogsConfig(valuesInput.Fluentd.Logs.Systemd),
			Kubelet:    createFluentdLogsConfig(valuesInput.Fluentd.Logs.Kubelet),
			Default:    createFluentdLogsConfig(valuesInput.Fluentd.Logs.Default),
		},
	}
}
