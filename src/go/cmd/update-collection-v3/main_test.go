package main

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndividualMigrations(t *testing.T) {
	_, testFileName, _, _ := runtime.Caller(0)
	currentTestDirectory := path.Dir(testFileName)
	for _, migration := range migrations {
		testMigrationsInDirectory(t, migration.action, path.Join(currentTestDirectory, "migrations", migration.directory, "testdata"))
	}
}

func TestAllMigrations(t *testing.T) {
	_, testFileName, _, _ := runtime.Caller(0)
	currentTestDirectory := path.Dir(testFileName)
	testMigrationsInDirectory(t, migrateYaml, path.Join(currentTestDirectory, "testdata"))
}

// testMigrationsInDirectory runs the migrate function on the *.input.yaml files
// in the directory and asserts that the output matches the contents of the *.output.yaml files.
func testMigrationsInDirectory(t *testing.T, migrate migrateFunc, directory string) {
	inputFileNames, err := filepath.Glob(path.Join(directory, "*.input.yaml"))
	require.NoError(t, err)
	for _, inputFileName := range inputFileNames {
		t.Run(path.Join(path.Base(path.Dir(directory)), path.Base(directory), path.Base(inputFileName)), func(t *testing.T) {
			inputFileContents, err := os.ReadFile(inputFileName)
			require.NoError(t, err)

			actualOutput, err := migrate(string(inputFileContents))
			require.NoError(t, err)

			outputFileName := strings.TrimSuffix(inputFileName, ".input.yaml") + ".output.yaml"
			expectedOutput, err := os.ReadFile(outputFileName)
			require.NoError(t, err)
			assert.Equal(t, string(expectedOutput), actualOutput)
		})
	}
}
