package tailingsidecaroperatorupgrade

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Values struct {
	TailingSidecarOperator struct {
		Rest map[string]interface{} `yaml:",inline"`
	} `yaml:"tailing-sidecar-operator,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
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

	return inputYaml, nil
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
	if len(values.TailingSidecarOperator.Rest) == 0 {
		return ""
	}

	return "WARNING! Changes in tailing-sidecar-operator detected, which may require manual migration\n" +
		"For details please see the following documentations:\n  - https://github.com/SumoLogic/sumologic-kubernetes-collection/blob/main/docs/v3-migration-doc.md\n" +
		"  - https://github.com/SumoLogic/tailing-sidecar/blob/63e7c7f38e9e1edf1a105407b4aea8322101ab8a/CHANGELOG.md"
}
