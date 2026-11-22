package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFailedInvocationExportsMetricsBeforeExit(t *testing.T) {
	dir := t.TempDir()
	telemetry := filepath.Join(dir, "telemetry.json")
	config := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(config, []byte(fmt.Sprintf("otel:\n  type: file\n  file:\n    filename: %q\n  sampler:\n    parentSampler: always\n", telemetry)), 0600))
	cmd := command()
	cmd.SetArgs([]string{"--config", config, "--retention-age", "0s"})
	require.ErrorContains(t, cmd.ExecuteContext(t.Context()), "must be positive")
	data, err := os.ReadFile(telemetry)
	require.NoError(t, err)
	require.NotContains(t, string(data), "runs.retention.invocations")
	require.Contains(t, string(data), "runs.retention.duration")
	require.Contains(t, string(data), "failure")
}
