package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	disablethanos "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/disable-thanos"
	"github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/events"
	eventsconfigmerge "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/events-config-merge"
	falcoupgrade "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/falco-upgrade"
	fluentdlogsconfigs "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/fluentd-logs-configs"
	kubeprometheusstackrepository "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/kube-prometheus-stack-repository"
	kubestatemetricscollectors "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/kube-state-metrics-collectors"
	"github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/logsmetadataconfig"
	metricsmetadataconfig "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/metrics-metadata-config"
	metricsserverupgrade "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/metrics-server-upgrade"
	otellogsconfigmerge "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/otellogs-config-merge"
	removeloadconfigfile "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/remove-load-config-file"
	tailingsidecaroperatorupgrade "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/tailing-sidecar-operator-upgrade"
	tracingreplaces "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/tracing-replaces"
	tracingconfig "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/tracing-config"
	tracingobjectchanges "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/tracing-objects-changes"
	"gopkg.in/yaml.v3"
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
		action: tracingobjectchanges.Migrate,
	},
	{
		directory: "tracing-config",
		action:    tracingconfig.Migrate,
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
	// TODO: order keys the same as input
	output, err = reorderYaml(output)

	return output, err
}

// reorder yaml keys in the input
// right now this just unmarshals into a map and marshals again
// the result is alphabetical key ordering
func reorderYaml(input string) (string, error) {
	var outputMap map[string]interface{}
	err := yaml.Unmarshal([]byte(input), &outputMap)
	if err != nil {
		return "", err
	}
	buffer := bytes.Buffer{}
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	err = encoder.Encode(outputMap)

	return buffer.String(), err
}

type migrateFunc func(string) (string, error)
