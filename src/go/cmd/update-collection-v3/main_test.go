package main

import (
	"testing"
	"strings"
	"github.com/stretchr/testify/require"
)

type TestCase struct {
	inputYaml string
	outputYaml string
	err error
	description string
}

func runYamlTest(t *testing.T, testCase TestCase) {
	actualOutput, err := migrateYaml(testCase.inputYaml)
	if testCase.err == nil {
		require.NoError(t, err, testCase.description)
	} else {
		require.Equal(t, err, testCase.err, testCase.description)
	}
	require.Equal(t, strings.Trim(testCase.outputYaml, "\n "), strings.Trim(actualOutput, "\n "), testCase.description)
}
