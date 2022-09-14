package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	disablethanos "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/disable-thanos"
	kubeprometheusstackandevents "github.com/SumoLogic/sumologic-kubernetes-collection/tools/cmd/update-collection-v3/migrations/kube-prometheus-stack-and-events"
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

func migrateYaml(input string) (string, error) {
	values, err := kubeprometheusstackandevents.Migrate(string(input))
	if err != nil {
		return "", fmt.Errorf("error running migration 'kube-prometheus-stack-and-events': %v", err)
	}

	values, err = disablethanos.Migrate(values)
	if err != nil {
		return "", fmt.Errorf("error running migration 'disable-thanos': %v", err)
	}

	return values, nil
}
