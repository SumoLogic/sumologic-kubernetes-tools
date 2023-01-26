package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	disablethanos "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/disable-thanos"
	"github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/events"
	eventsconfigmerge "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/events-config-merge"
	falcoupgrade "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/falco-upgrade"
	fluentdautoscaling "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/fluentd-autoscaling"
	fluentdlogsconfigs "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/fluentd-logs-configs"
	kubeprometheusstackrepository "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/kube-prometheus-stack-repository"
	kubestatemetricscollectors "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/kube-state-metrics-collectors"
	"github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/logformat"
	logsmetadataconfig "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/logs-metadata-config"
	metricsmetadataconfig "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/metrics-metadata-config"
	metricsserverupgrade "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/metrics-server-upgrade"
	otellogsconfigmerge "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/otellogs-config-merge"
	removeloadconfigfile "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/remove-load-config-file"
	tailingsidecaroperatorupgrade "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/tailing-sidecar-operator-upgrade"
	tracingconfig "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/tracing-config"
	tracingobjectchanges "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/tracing-objects-changes"
	tracingreplaces "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/tracing-replaces"
	"github.com/goccy/go-yaml"
)

var (
	inFileFlag  = flag.String("in", "values.yaml", "input values.yaml to be migrated")
	outFileFlag = flag.String("out", "new_values.yaml", "output values.yaml")
)

type Migration struct {
	directory string
	action    migrateFunc
}

func main() {
	flag.Parse()

	err := migrateYamlFile(*inFileFlag, *outFileFlag)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Successfully migrated the configuration")
}

func migrateYamlFile(yamlV2FilePath string, yamlV3FilePath string) error {
	f, err := os.Open(yamlV2FilePath)
	if err != nil {
		return fmt.Errorf("cannot open file %s: %v", *inFileFlag, err)
	}

	yamlV2, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("error reading from file %s: %v", yamlV2FilePath, err)
	}

	yamlV3, err := migrateYaml(string(yamlV2))
	if err != nil {
		return fmt.Errorf("error migrating values %v", err)
	}

	err = os.WriteFile(yamlV3FilePath, []byte(yamlV3), 0666)
	if err != nil {
		return fmt.Errorf("failed writing %s: %v", *outFileFlag, err)
	}

	return nil
}

var migrations = []Migration{
	{
		directory: "kube-prometheus-stack",
		action:    kubestatemetricscollectors.Migrate,
	},
	{
		directory: "events",
		action:    events.Migrate,
	},
	{
		directory: "disable-thanos",
		action:    disablethanos.Migrate,
	},
	{
		directory: "remove-load-config-file",
		action:    removeloadconfigfile.Migrate,
	},
	{
		directory: "tracing-replaces",
		action:    tracingreplaces.Migrate,
	},
	{
		directory: "falco-upgrade",
		action:    falcoupgrade.Migrate,
	},
	{
		directory: "events-config-merge",
		action:    eventsconfigmerge.Migrate,
	},
	{
		directory: "logs-metadata-config",
		action:    logsmetadataconfig.Migrate,
	},
	{
		directory: "metrics-metadata-config",
		action:    metricsmetadataconfig.Migrate,
	},
	{
		directory: "otellogs-config-merge",
		action:    otellogsconfigmerge.Migrate,
	},
	{
		directory: "kube-prometheus-stack-repository",
		action:    kubeprometheusstackrepository.Migrate,
	},
	{
		directory: "metrics-server-upgrade",
		action:    metricsserverupgrade.Migrate,
	},
	{
		directory: "tailing-sidecar-operator-upgrade",
		action:    tailingsidecaroperatorupgrade.Migrate,
	},
	{
		directory: "fluentd-logs-configs",
		action:    fluentdlogsconfigs.Migrate,
	},
	{
		directory: "tracing-objects-changes",
		action:    tracingobjectchanges.Migrate,
	},
	{
		directory: "tracing-config",
		action:    tracingconfig.Migrate,
	},
	{
		directory: "fluentd-autoscaling",
		action:    fluentdautoscaling.Migrate,
	},
	{
		directory: "logformat",
		action:    logformat.Migrate,
	},
}

func migrateYaml(input string) (string, error) {
	var err error
	output := input

	for _, migration := range migrations {
		output, err = migration.action(output)
		if err != nil {
			return "", fmt.Errorf("error running migration '%s': %v", migration.directory, err)
		}
	}

	// make the output consistently ordered
	// without this logic, key ordering would depend on the final migration
	output, err = reorderYaml(output, input)

	return output, err
}

// reorder yaml keys in the input based on their order in the output
func reorderYaml(input string, original string) (string, error) {
	var outputMapSlice, originalMapSlice yaml.MapSlice
	var err error
	err = yaml.UnmarshalWithOptions([]byte(input), &outputMapSlice, yaml.UseOrderedMap())
	if err != nil {
		return "", err
	}
	err = yaml.UnmarshalWithOptions([]byte(original), &originalMapSlice, yaml.UseOrderedMap())
	if err != nil {
		return "", err
	}

	sortByBlueprint(outputMapSlice, originalMapSlice)

	outputBytes, err := yaml.MarshalWithOptions(outputMapSlice, yaml.Indent(2), yaml.UseLiteralStyleIfMultiline(true))

	return string(outputBytes), err
}

// sortByBlueprint sorts the input based on the order of keys in the output
// this is done recursively
// keys not present in the blueprint go to the end and are sorted alphabetically
func sortByBlueprint(input yaml.MapSlice, blueprint yaml.MapSlice) {
	blueprintMap := blueprint.ToMap()
	sort.Slice(input, func(i, j int) bool {
		iKey := input[i].Key
		jKey := input[j].Key
		iPosition, jPosition := len(input), len(input)
		for position, entry := range blueprint {
			if entry.Key == iKey {
				iPosition = position
			}
			if entry.Key == jKey {
				jPosition = position
			}
		}
		if iPosition == len(input) && jPosition == len(input) {
			// if neither key are in the blueprint, sort alphabetically
			return iKey.(string) < jKey.(string)
		}
		return iPosition < jPosition
	})

	// sort recursively for values which are also yaml.MapSlice and exist in the blueprint
	var entryValueMapSlice yaml.MapSlice
	var ok bool
	for _, entry := range input {
		if entryValueMapSlice, ok = (entry.Value).(yaml.MapSlice); !ok { // not a yaml.MapSlice
			continue
		}
		if blueprintValue, ok := blueprintMap[entry.Key]; ok {
			if blueprintValueMapSlice, ok := (blueprintValue).(yaml.MapSlice); ok {
				sortByBlueprint(entryValueMapSlice, blueprintValueMapSlice)
			}
		}
	}
}

type migrateFunc func(string) (string, error)
