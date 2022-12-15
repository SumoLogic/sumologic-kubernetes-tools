package fluentdlogsconfigs

type ContainersLogsConfig struct {
	SourceName                     *string                         `yaml:"sourceName,omitempty"`
	SourceCategory                 *string                         `yaml:"sourceCategory,omitempty"`
	SourceCategoryPrefix           *string                         `yaml:"sourceCategoryPrefix,omitempty"`
	SourceCategoryReplaceDash      *string                         `yaml:"sourceCategoryReplaceDash,omitempty"`
	ExcludeFacilityRegex           *string                         `yaml:"excludeFacilityRegex,omitempty"`
	ExcludeHostRegex               *string                         `yaml:"excludeHostRegex,omitempty"`
	ExcludePriorityRegex           *string                         `yaml:"excludePriorityRegex,omitempty"`
	ExcludeUnitRegex               *string                         `yaml:"excludeUnitRegex,omitempty"`
	PerContainerAnnotationsEnabled *bool                           `yaml:"perContainerAnnotationsEnabled,omitempty"`
	PerContainerAnnotationPrefixes *PerContainerAnnotationPrefixes `yaml:"perContainerAnnotationPrefixes,omitempty"`
	Rest                           map[string]interface{}          `yaml:",inline"`
}

type LogsConfig struct {
	SourceName                *string                `yaml:"sourceName,omitempty"`
	SourceCategory            *string                `yaml:"sourceCategory,omitempty"`
	SourceCategoryPrefix      *string                `yaml:"sourceCategoryPrefix,omitempty"`
	SourceCategoryReplaceDash *string                `yaml:"sourceCategoryReplaceDash,omitempty"`
	ExcludeFacilityRegex      *string                `yaml:"excludeFacilityRegex,omitempty"`
	ExcludeHostRegex          *string                `yaml:"excludeHostRegex,omitempty"`
	ExcludePriorityRegex      *string                `yaml:"excludePriorityRegex,omitempty"`
	ExcludeUnitRegex          *string                `yaml:"excludeUnitRegex,omitempty"`
	Rest                      map[string]interface{} `yaml:",inline"`
}

type Fluentd struct {
	Logs *FluentdLogs           `yaml:"logs,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}

type FluentdLogs struct {
	Containers *ContainersLogsConfig  `yaml:"containers,omitempty"`
	Systemd    *LogsConfig            `yaml:"systemd,omitempty"`
	Kubelet    *LogsConfig            `yaml:"kubelet,omitempty"`
	Default    *LogsConfig            `yaml:"default,omitempty"`
	Rest       map[string]interface{} `yaml:",inline"`
}

type PerContainerAnnotationPrefixes []string
