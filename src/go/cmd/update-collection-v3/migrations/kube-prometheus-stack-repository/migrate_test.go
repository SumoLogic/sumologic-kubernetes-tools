package kubeprometheusstackrepository

import (
	"testing"

	"gotest.tools/assert"
)

var (
	link interface{} = "link"
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
				values.KubePrometheusStack.AlertManager.AlertManagerSpec.Image.Repository = &link
				values.KubePrometheusStack.PrometheusOperator.AdmissionWebhooks.Patch.Image.Repository = &link
				values.KubePrometheusStack.PrometheusOperator.Image.Repository = &link
				values.KubePrometheusStack.PrometheusOperator.PrometheusConfigReloader.Image.Repository = &link
				values.KubePrometheusStack.PrometheusOperator.ThanosImage.Repository = &link
				values.KubePrometheusStack.Prometheus.PrometheusSpec.Image.Repository = &link
				values.KubePrometheusStack.ThanosRuler.ThanosRulerSpec.Image.Repository = &link
				return values
			},
			expected: `WARNING! Found following values in kube-prometheus-stack configuration which must be manually migrated:
alertmanager.alertmanagerSpec.image.repository
prometheus.prometheusSpec.image.repository
prometheusOperator.admissionWebhooks.patch.image.repository
prometheusOperator.image.repository
prometheusOperator.prometheusConfigReloader.image.repository
prometheusOperator.thanosImage.repository
thanosRuler.thanosRulerSpec.image.repository
for details please see the following documentations:
  - https://github.com/SumoLogic/sumologic-kubernetes-collection/blob/main/docs/v3-migration-doc.md
  - https://github.com/prometheus-community/helm-charts/tree/kube-prometheus-stack-42.1.0/charts/kube-prometheus-stack#from-41x-to-42x`,
		},
		{
			name: "partial",
			input: func() *Values {
				values := &Values{}
				values.KubePrometheusStack.PrometheusOperator.AdmissionWebhooks.Patch.Image.Repository = &link
				values.KubePrometheusStack.PrometheusOperator.PrometheusConfigReloader.Image.Repository = &link
				values.KubePrometheusStack.PrometheusOperator.ThanosImage.Repository = &link
				values.KubePrometheusStack.ThanosRuler.ThanosRulerSpec.Image.Repository = &link
				return values
			},
			expected: `WARNING! Found following values in kube-prometheus-stack configuration which must be manually migrated:
prometheusOperator.admissionWebhooks.patch.image.repository
prometheusOperator.prometheusConfigReloader.image.repository
prometheusOperator.thanosImage.repository
thanosRuler.thanosRulerSpec.image.repository
for details please see the following documentations:
  - https://github.com/SumoLogic/sumologic-kubernetes-collection/blob/main/docs/v3-migration-doc.md
  - https://github.com/prometheus-community/helm-charts/tree/kube-prometheus-stack-42.1.0/charts/kube-prometheus-stack#from-41x-to-42x`,
		},
		{
			name:     "no values",
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
