package kubeprometheusstackrepository

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Values struct {
	KubePrometheusStack struct {
		AlertManager struct {
			AlertManagerSpec struct {
				Image Image                  `yaml:"image,omitempty"`
				Rest  map[string]interface{} `yaml:",inline"`
			} `yaml:"alertmanagerSpec,omitempty"`
			Rest map[string]interface{} `yaml:",inline"`
		} `yaml:"alertmanager,omitempty"`
		PrometheusOperator struct {
			AdmissionWebhooks struct {
				Patch struct {
					Image Image                  `yaml:"image,omitempty"`
					Rest  map[string]interface{} `yaml:",inline"`
				} `yaml:"patch,omitempty"`
			} `yaml:"admissionWebhooks,omitempty"`
			Rest                     map[string]interface{} `yaml:",inline"`
			Image                    Image                  `yaml:"image,omitempty"`
			PrometheusConfigReloader struct {
				Image Image                  `yaml:"image,omitempty"`
				Rest  map[string]interface{} `yaml:",inline"`
			} `yaml:"prometheusConfigReloader,omitempty"`
			ThanosImage struct {
				Repository *interface{}           `yaml:"repository,omitempty"`
				Rest       map[string]interface{} `yaml:",inline"`
			} `yaml:"thanosImage,omitempty"`
		} `yaml:"prometheusOperator,omitempty"`
		Prometheus struct {
			PrometheusSpec struct {
				Image Image                  `yaml:"image,omitempty"`
				Rest  map[string]interface{} `yaml:",inline"`
			} `yaml:"prometheusSpec,omitempty"`
			Rest map[string]interface{} `yaml:",inline"`
		} `yaml:"prometheus,omitempty"`
		ThanosRuler struct {
			ThanosRulerSpec struct {
				Image Image                  `yaml:"image,omitempty"`
				Rest  map[string]interface{} `yaml:",inline"`
			} `yaml:"thanosRulerSpec,omitempty"`
			Rest map[string]interface{} `yaml:",inline"`
		} `yaml:"thanosRuler,omitempty"`
		Rest map[string]interface{} `yaml:",inline"`
	} `yaml:"kube-prometheus-stack,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}

type Image struct {
	Repository *interface{}           `yaml:"repository,omitempty"`
	Rest       map[string]interface{} `yaml:",inline"`
}

func Migrate(inputYaml string) (outputYaml string, err error) {
	values, err := parseValues(inputYaml)
	if err != nil {
		return "", fmt.Errorf("error parsing input yaml: %v", err)
	}

	values = migrate(&values)
	if err != nil {
		return "", fmt.Errorf("error migrating: %v", err)
	}

	buffer := bytes.Buffer{}
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	err = encoder.Encode(values)
	return buffer.String(), err
}

func parseValues(inputYaml string) (Values, error) {
	var v Values
	err := yaml.Unmarshal([]byte(inputYaml), &v)
	return v, err
}

func migrate(values *Values) Values {
	log := migrateLog(values)

	if log != "" {
		fmt.Println(log)
	}

	return *values
}

func migrateLog(values *Values) string {
	migrationMap := map[string]*interface{}{
		"alertmanager.alertmanagerSpec.image.repository":               values.KubePrometheusStack.AlertManager.AlertManagerSpec.Image.Repository,
		"prometheusOperator.admissionWebhooks.patch.image.repository":  values.KubePrometheusStack.PrometheusOperator.AdmissionWebhooks.Patch.Image.Repository,
		"prometheusOperator.image.repository":                          values.KubePrometheusStack.PrometheusOperator.Image.Repository,
		"prometheusOperator.prometheusConfigReloader.image.repository": values.KubePrometheusStack.PrometheusOperator.PrometheusConfigReloader.Image.Repository,
		"prometheusOperator.thanosImage.repository":                    values.KubePrometheusStack.PrometheusOperator.ThanosImage.Repository,
		"prometheus.prometheusSpec.image.repository":                   values.KubePrometheusStack.Prometheus.PrometheusSpec.Image.Repository,
		"thanosRuler.thanosRulerSpec.image.repository":                 values.KubePrometheusStack.ThanosRuler.ThanosRulerSpec.Image.Repository,
	}
	toMigrate := []string{}
	for k, v := range migrationMap {
		if v != nil {
			toMigrate = append(toMigrate, k)
		}
	}

	sort.Strings(toMigrate)

	if len(toMigrate) == 0 {
		return ""
	}

	return "WARNING! Found following values in kube-prometheus-stack configuration which must be manually migrated:\n" +
		strings.Join(toMigrate, "\n") +
		"\nfor details please see the following documentations:\n  - https://github.com/SumoLogic/sumologic-kubernetes-collection/blob/main/docs/v3-migration-doc.md\n" +
		"  - https://github.com/prometheus-community/helm-charts/tree/kube-prometheus-stack-42.1.0/charts/kube-prometheus-stack#from-41x-to-42x"
}
