package main

import (
	"bufio"
	"flag"
	"fmt"
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

	valuesV2, err := parseValues(*inFileFlag)
	if err != nil {
		log.Fatalf("failed reading %s: %v", *inFileFlag, err)
	}

	if err := migrate(&valuesV2); err != nil {
		log.Fatalf("failed migrating %s: %v", *inFileFlag, err)
	}

	if err := toYaml(valuesV2, *outFileFlag); err != nil {
		log.Fatalf("failed writing %s: %v", *outFileFlag, err)
	}
}

func parseValues(path string) (ValuesV2, error) {
	f, err := os.Open(path)
	if err != nil {
		return ValuesV2{}, fmt.Errorf("cannot open file %s: %w", path, err)
	}

	decoder := yaml.NewDecoder(bufio.NewReader(f))

	var valuesV2 ValuesV2
	if err := decoder.Decode(&valuesV2); err != nil {
		return ValuesV2{}, fmt.Errorf("cannot unmarshal data from %s: %w", path, err)
	}

	return valuesV2, nil
}

func toYaml(valuesV2 ValuesV2, path string) error {
	out, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("cannot open new file for writing (%s): %v", path, err)
	}

	enc := yaml.NewEncoder(out)
	enc.SetIndent(2)
	defer func() {
		if errClose := enc.Close(); errClose != nil {
			log.Fatalf("failed closing yaml encoder (%s): %v", path, err)
		}
	}()

	if err := enc.Encode(valuesV2); err != nil {
		return fmt.Errorf("failed writing new values.yaml (%s): %v", path, err)
	}

	return nil
}
