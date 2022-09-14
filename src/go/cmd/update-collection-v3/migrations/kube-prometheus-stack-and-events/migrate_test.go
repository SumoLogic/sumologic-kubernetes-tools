package kubeprometheusstackandevents

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrateKubeStateMetricsCollectors(t *testing.T) {
	for _, test := range []struct {
		Name     string
		V2Config *map[string]bool
		V3Config *[]string
	}{
		{
			Name:     "nil",
			V2Config: nil,
			V3Config: nil,
		},
		{
			Name: "disable some collectors",
			V2Config: &map[string]bool{
				"certificatesigningrequests": false,
				"deployments":                false,
				"configmaps":                 true,
				"poddisruptionbudgets":       false,
			},
			V3Config: &[]string{
				"configmaps",
				"cronjobs",
				"daemonsets",
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
				"pods",
				"replicasets",
				"replicationcontrollers",
				"resourcequotas",
				"secrets",
				"services",
				"statefulsets",
				"storageclasses",
				"validatingwebhookconfigurations",
				"volumeattachments"},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			result := migrateKubeStateMetricsCollectors(test.V2Config)
			assert.Equal(t, test.V3Config, result)
		})
	}
}
