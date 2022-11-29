package metricsserverupgrade

import (
	"testing"

	"gotest.tools/assert"
)

var (
	foo interface{} = "bar"
)

func TestMain(t *testing.T) {
	for _, tt := range []struct {
		input    func() *Values
		expected string
		name     string
	}{
		{
			name: "all",
			input: func() *Values {
				values := &Values{}
				values.MetricsServer.Rest = map[string]interface{}{
					"foo": foo,
				}
				return values
			},
			expected: `WARNING! Changes in metrics-server detected, which may require manual migration
For details please see the following documentations:
  - https://github.com/SumoLogic/sumologic-kubernetes-collection/blob/main/docs/v3-migration-doc.md
  - https://github.com/bitnami/charts/tree/5b09f7a7c0d9232f5752840b6c4e5cdc56d7f796/bitnami/metrics-server#to-600`,
		},
		{
			name: "no values",
			input: func() *Values {
				values := &Values{}
				return values
			},
			expected: ``,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			val := migrateLog(tt.input())
			assert.Equal(t, tt.expected, val)
		})
	}
}
