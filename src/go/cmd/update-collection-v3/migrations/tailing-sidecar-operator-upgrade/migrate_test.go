package tailingsidecaroperatorupgrade

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
				values.TailingSidecarOperator.Rest = map[string]interface{}{
					"foo": foo,
				}
				return values
			},
			expected: `WARNING! Changes in tailing-sidecar-operator detected, which may require manual migration
For details please see the following documentations:
  - https://github.com/SumoLogic/sumologic-kubernetes-collection/blob/main/docs/v3-migration-doc.md
  - https://github.com/SumoLogic/tailing-sidecar/blob/63e7c7f38e9e1edf1a105407b4aea8322101ab8a/CHANGELOG.md`,
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
