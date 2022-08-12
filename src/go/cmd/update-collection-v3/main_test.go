package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestCase struct {
	inputYaml   string
	outputYaml  string
	err         error
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

func TestYaml(t *testing.T) {
	for _, tt := range []TestCase{
		{
			inputYaml: `
kube-prometheus-stack:
  kube-state-metrics:
    collectors:
      certificatesigningrequests: false
      configmaps: true
      persistentvolumes: false`,
			outputYaml: `
kube-prometheus-stack:
  kube-state-metrics:
    collectors:
      - configmaps
      - cronjobs
      - daemonsets
      - deployments
      - endpoints
      - horizontalpodautoscalers
      - ingresses
      - jobs
      - limitranges
      - mutatingwebhookconfigurations
      - namespaces
      - networkpolicies
      - nodes
      - persistentvolumeclaims
      - poddisruptionbudgets
      - pods
      - replicasets
      - replicationcontrollers
      - resourcequotas
      - secrets
      - services
      - statefulsets
      - storageclasses
      - validatingwebhookconfigurations
      - volumeattachments
`,
			err:         nil,
			description: "kube state metrics migration",
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			runYamlTest(t, tt)
		})
	}
}
