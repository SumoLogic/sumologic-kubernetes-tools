package disablethanos

import (
	"io/ioutil"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisableThanos(t *testing.T) {
	_, testFileName, _, _ := runtime.Caller(0)
	testDir := path.Dir(testFileName)
	inputFileNames, err := filepath.Glob(path.Join(testDir, "testdata", "*.input.yaml"))
	require.NoError(t, err)
	for _, inputFileName := range inputFileNames {
		t.Run(path.Base(inputFileName), func(t *testing.T) {
			inputFileContents, err := ioutil.ReadFile(inputFileName)
			require.NoError(t, err)

			outputFileContents, err := Migrate(string(inputFileContents))
			require.NoError(t, err)

			expectedOutputFileName := strings.TrimSuffix(inputFileName, ".input.yaml") + ".output.yaml"
			expectedOutputFileContents, err := ioutil.ReadFile(expectedOutputFileName)
			require.NoError(t, err)
			assert.Equal(t, string(expectedOutputFileContents), outputFileContents)
		})
	}
}
