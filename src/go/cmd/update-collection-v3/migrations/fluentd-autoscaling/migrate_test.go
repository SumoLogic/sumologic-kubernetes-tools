package fluentdautoscaling

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestCase struct {
	inputYaml   string
	outputYaml  string
	err         error
	description string
}

func runYamlTest(t *testing.T, testCase TestCase) {
	actualOutput, err := Migrate(testCase.inputYaml)
	if testCase.err == nil {
		require.NoError(t, err, testCase.description)
	} else {
		require.Equal(t, err, testCase.err, testCase.description)
	}
	require.Equal(t, strings.Trim(testCase.outputYaml, "\n "), strings.Trim(actualOutput, "\n "), testCase.description)
}

func Test_EventsEnabledProvider(t *testing.T) {
	testCases := []TestCase{
		{
			inputYaml:   `{}`,
			outputYaml:  `{}`,
			description: "No config, no change",
		},
		{
			inputYaml: `
fluentd:
  key: value
`,
			outputYaml: `
fluentd:
  key: value
`,
			description: "No config, no change",
		},
		{
			inputYaml: `
fluentd:
  logs:
    key: value
  metrics:
    key: value
`,
			outputYaml: `
fluentd:
  logs:
    key: value
  metrics:
    key: value
`,
			description: "No config, no change",
		},
		{
			inputYaml: `
fluentd:
  logs:
    autoscaling:
      minReplicas: 5
  metrics:
    autoscaling:
      minReplicas: 5
`,
			outputYaml: `
fluentd:
  logs:
    autoscaling:
      minReplicas: 5
  metrics:
    autoscaling:
      minReplicas: 5
`,
			description: "No config, no change",
		},
		{
			inputYaml: `
fluentd:
  logs:
    autoscaling:
      enabled: true
      minReplicas: 5
  metrics:
    autoscaling:
      enabled: true
      minReplicas: 5
`,
			outputYaml: `
metadata:
  logs:
    autoscaling:
      enabled: true
  metrics:
    autoscaling:
      enabled: true
fluentd:
  logs:
    autoscaling:
      enabled: true
      minReplicas: 5
  metrics:
    autoscaling:
      enabled: true
      minReplicas: 5
`,
			description: "Enabled, everything else unchanged",
		},
		{
			inputYaml: `
metadata:
  logs:
    autoscaling:
      enabled: false
  metrics:
    autoscaling:
      enabled: false
fluentd:
  logs:
    autoscaling:
      enabled: true
      minReplicas: 5
  metrics:
    autoscaling:
      enabled: true
      minReplicas: 5
`,
			outputYaml: `
metadata:
  logs:
    autoscaling:
      enabled: false
  metrics:
    autoscaling:
      enabled: false
fluentd:
  logs:
    autoscaling:
      enabled: true
      minReplicas: 5
  metrics:
    autoscaling:
      enabled: true
      minReplicas: 5
`,
			description: "Do nothing if metadata autoscaling already disabled",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			runYamlTest(t, testCase)
		})
	}
}
