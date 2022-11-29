package falcoupgrade

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

func Migrate(input string) (string, error) {
	values, err := parseValues(input)
	if err != nil {
		return "", fmt.Errorf("error parsing input yaml: %v", err)
	}

	if values.Falco.Enabled != nil {
		fmt.Println(`WARNING! Found that falco configuration is/was enabled. Performing automatic migration of default keys.
Please confirm that migrated configuration is correct according to Falco helm chart: https://github.com/falcosecurity/charts/tree/falco-2.4.2/falco`)

		if values.Falco.Image != nil {
			fmt.Println("Please migrate falco.image manually.\n" +
				"  - Default for falco.image.repository is now falcosecurity/falco-no-driver\n" +
				"  - Default for falco.driver.loader.initContainer.image.repository is now falcosecurity/falco-driver-loader")
		}

		if values.Falco.ExtraInitContainers != nil {
			fmt.Println("Renaming falco.falco.extraInitContainers to falco.extra.initContainers")
			if values.Falco.Extra.InitContainers != nil {
				fmt.Println(`WARNING! falco.falco.extraInitContainers already set. Please migrate falco.falco.jsonOutput manually`)
			} else {
				values.Falco.Extra.InitContainers = values.Falco.ExtraInitContainers
				values.Falco.ExtraInitContainers = nil
			}
		}

		if values.Falco.Falco.JsonOutputOld != nil {
			fmt.Println("Renaming falco.falco.jsonOutput to falco.falco.json_output")
			if values.Falco.Falco.JsonOutputNew != nil {
				fmt.Println(`WARNING! falco.falco.json_output already set. Please migrate falco.falco.jsonOutput manually`)
			} else {
				values.Falco.Falco.JsonOutputNew = values.Falco.Falco.JsonOutputOld
				values.Falco.Falco.JsonOutputOld = nil
			}
		}

		if values.Falco.Falco.RulesFileOld != nil {
			fmt.Println("Renaming falco.falco.rulesFile to falco.falco.rules_file")
			if values.Falco.Falco.RulesFileNew != nil {
				fmt.Println(`WARNING! falco.falco.rules_file already set. Please migrate falco.falco.rulesFile manually`)
			} else {
				values.Falco.Falco.RulesFileNew = values.Falco.Falco.RulesFileOld
				values.Falco.Falco.RulesFileOld = nil
			}
		}

		ebpf := "ebpf"
		if values.Falco.Ebpf != nil && values.Falco.Ebpf.Enabled != nil && *values.Falco.Ebpf.Enabled {
			fmt.Println("Setting falco.driver.kind to `ebpf` as `falco.ebpf.enabled` is set to `true`")
			values.Falco.Driver = &struct {
				Kind *string "yaml:\"kind,omitempty\""
			}{Kind: &ebpf}

			if len(values.Falco.Ebpf.Rest) == 0 {
				values.Falco.Ebpf = nil
			} else {
				values.Falco.Ebpf.Enabled = nil
			}
		}
	}

	buffer := bytes.Buffer{}
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	err = encoder.Encode(values)
	return buffer.String(), err
}

func parseValues(input string) (Values, error) {
	var v Values
	err := yaml.Unmarshal([]byte(input), &v)
	return v, err
}
