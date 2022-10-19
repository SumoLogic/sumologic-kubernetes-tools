package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	disablethanos "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/disable-thanos"
	"github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/events"
	kubeprometheusstack "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/kube-prometheus-stack"
)

var (
	inFileFlag  = flag.String("in", "values.yaml", "input values.yaml to be migrated")
	outFileFlag = flag.String("out", "new_values.yaml", "output values.yaml")
)

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

	yamlV2, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("error reading from file %s: %v", yamlV2FilePath, err)
	}

	yamlV3, err := migrateYaml(string(yamlV2))
	if err != nil {
		return fmt.Errorf("error migrating values %v", err)
	}

	err = ioutil.WriteFile(yamlV3FilePath, []byte(yamlV3), 0666)
	if err != nil {
		return fmt.Errorf("failed writing %s: %v", *outFileFlag, err)
	}

	return nil
}

var migrationDirectoriesAndFunctions = map[string]migrateFunc{
	"kube-prometheus-stack": kubeprometheusstack.Migrate,
	"events":                events.Migrate,
	"disable-thanos":        disablethanos.Migrate,
}

func migrateYaml(input string) (string, error) {
	output := input

	for migrationDir, migrate := range migrationDirectoriesAndFunctions {
		var err error
		output, err = migrate(output)
		if err != nil {
			return "", fmt.Errorf("error running migration '%s': %v", migrationDir, err)
		}
	}

	return output, nil
}

type migrateFunc func(string) (string, error)
