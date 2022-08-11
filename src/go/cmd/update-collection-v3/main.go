package main

import (
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v3"
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

	err = ioutil.WriteFile(yamlV3FilePath, []byte(yamlV3), fs.ModeType)
	if err != nil {
		return fmt.Errorf("failed writing %s: %v", *outFileFlag, err)
	}

	return nil
}

func migrateYaml(yamlV2 string) (yamlV3 string, err error) {
	valuesV2, err := parseValues(yamlV2)
	if err != nil {
		return "", fmt.Errorf("failed reading %s: %v", *inFileFlag, err)
	}

	valuesV3, err := migrate(&valuesV2)
	if err != nil {
		return "", fmt.Errorf("failed migrating %s: %v", *inFileFlag, err)
	}

	yamlV3Bytes, err := yaml.Marshal(valuesV3)
	return string(yamlV3Bytes), err
}

func parseValues(yamlV2 string) (ValuesV2, error) {
	var valuesV2 ValuesV2
	err := yaml.Unmarshal([]byte(yamlV2), &valuesV2)
	return valuesV2, err
}
