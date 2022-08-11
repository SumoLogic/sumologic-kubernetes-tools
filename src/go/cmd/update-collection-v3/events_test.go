package main

import (
	"testing"
)

func Test_EventsEnabledProvider(t *testing.T) {
	testCases := []TestCase{
		{
			inputYaml: `
sumologic:
  events:
    enabled: true
`,
			outputYaml:  `{}`,
			description: "Events are enabled by default",
		},
		{
			inputYaml: `
sumologic:
  events:
    enabled: false
`,
			outputYaml: `
sumologic:
  events:
    enabled: false
`,
			description: "Migration keeps events disabled",
		},
		{
			inputYaml: `
fluentd:
  events:
    enabled: false
`,
			outputYaml: `
sumologic:
  events:
    enabled: false
`,
			description: "Migration keeps events disabled for FluentD",
		},
		{
			inputYaml: `
sumologic:
  events:
    enabled: true
    provider: otelcol
`,
			outputYaml:  `{}`,
			description: "Default event provider is otelcol",
		},
		{
			inputYaml: `
sumologic:
  events:
    provider: fluentd
`,
			outputYaml: `
sumologic:
  events:
    provider: fluentd
`,
			description: "Keep FluentD as event provider if explicitly set",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			runYamlTest(t, testCase)
		})
	}
}

func Test_EventsSourceMetadata(t *testing.T) {
	testCases := []TestCase{
		{
			inputYaml: `
fluentd:
  events:
    sourceName: events
    sourceCategory: k8s/events
`,
			outputYaml: `
sumologic:
  events:
    sourceName: events
    sourceCategory: k8s/events
`,
			description: "Copy sourceName and sourceCategory from FluentD Events",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			runYamlTest(t, testCase)
		})
	}
}

func Test_EventsPersistence(t *testing.T) {
	testCases := []TestCase{
		{
			inputYaml: `
otelevents:
  persistence:
    enabled: true
    size: 10Gi
    accessMode: ReadWriteOnce
    pvcLabels:
      key: value
`,
			outputYaml: `
sumologic:
  events:
    persistence:
      enabled: true
      size: 10Gi
      persistentVolume:
        accessMode: ReadWriteOnce
        pvcLabels:
          key: value
`,
			description: "Use otelevents settings by default, if set",
		},
		{
			inputYaml: `
sumologic:
  events:
    provider: fluentd
fluentd:
  persistence:
    enabled: true
    size: 10Gi
    accessMode: ReadWriteOnce
    storageClass: default
`,
			outputYaml: `
sumologic:
  events:
    provider: fluentd
    persistence:
      enabled: true
      size: 10Gi
      persistentVolume:
        accessMode: ReadWriteOnce
        storageClass: default
fluentd:
  persistence:
    enabled: true
    size: 10Gi
    accessMode: ReadWriteOnce
    storageClass: default
`,
			description: "Use FluentD settings if it's explicitly set as provider",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			runYamlTest(t, testCase)
		})
	}
}
