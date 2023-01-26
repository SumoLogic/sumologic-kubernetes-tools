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
		migrationDirectory := path.Join(currentTestDirectory, "migrations", migration.directory)
		_, err := os.Stat(migrationDirectory)
		require.NoError(t, err, "migration directory '%s' not found", migration.directory)
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

func TestReorderYaml(t *testing.T) {
	testCases := []struct {
		inputYaml    string
		originalYaml string
		outputYaml   string
		description  string
	}{
		{
			inputYaml: `
a: b
c: d
`,
			originalYaml: `
c: d
a: b
`,
			outputYaml: `
c: d
a: b
`,
			description: "basic ordering",
		},
		{
			inputYaml: `
a: b
c: d
e: f
`,
			originalYaml: `
c: d
a: b
`,
			outputYaml: `
c: d
a: b
e: f
`,
			description: "key not in original goes to the end",
		},
		{
			inputYaml: `
g: h
e: f
`,
			originalYaml: `
c: d
a: b
`,
			outputYaml: `
e: f
g: h
`,
			description: "keys not in original are sorted alphabetically",
		},
		{
			inputYaml: `
nested: 
  a: b
  c: d
`,
			originalYaml: `
nested:
  c: d
  a: b
`,
			outputYaml: `
nested:
  c: d
  a: b
`,
			description: "basic nested ordering",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			actualOutput, err := reorderYaml(testCase.inputYaml, testCase.originalYaml)
			require.NoError(t, err)
			require.Equal(t, strings.Trim(testCase.outputYaml, "\n"), strings.Trim(actualOutput, "\n"))
		})
	}
}
