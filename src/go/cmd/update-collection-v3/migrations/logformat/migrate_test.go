package logformat

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
`,
			outputYaml: `
fluentd:
  logs:
    key: value
`,
			description: "No config, no change",
		},
		{
			inputYaml: `
fluentd:
  logs:
    output:
      key: value 
`,
			outputYaml: `
fluentd:
  logs:
    output:
      key: value 
`,
			description: "No config, no change",
		},
		{
			inputYaml: `
fluentd:
  logs:
    output:
      logFormat: text
`,
			outputYaml: `
sumologic:
  logs:
    container:
      format: text
`,
			description: "Moved",
		},
		{
			inputYaml: `
fluentd:
  logs:
    output:
      logFormat: text
sumologic:
  logs:
    container:
      format: json_merge
`,
			outputYaml: `
sumologic:
  logs:
    container:
      format: json_merge
`,
			description: "Don't overwrite sumologic format if already set",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			runYamlTest(t, testCase)
		})
	}
}
