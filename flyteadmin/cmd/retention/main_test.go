package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestFailedInvocationExportsMetrics(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "metrics.json")
	config := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(config, []byte(fmt.Sprintf("otel:\n  type: file\n  file:\n    filename: %q\n", output)), 0600))
	cmd := command()
	cmd.SetArgs([]string{"--config", config, "--retention-age", "0s"})
	require.ErrorContains(t, cmd.ExecuteContext(t.Context()), "must be positive")
	bytes, err := os.ReadFile(output)
	require.NoError(t, err)
	require.Contains(t, string(bytes), "flyteadmin.retention.duration")
	require.Contains(t, string(bytes), "failure")
	require.NotContains(t, string(bytes), "retention.invocations")
}
func TestQuoteDSN(t *testing.T) {
	for _, value := range []string{"simple", "has spaces", "has'quote", `has\slash`, "has' quote\\"} {
		cfg, err := pgx.ParseConfig("host=db.example.com sslmode=verify-full password=" + quoteDSN(value))
		require.NoError(t, err)
		require.Equal(t, value, cfg.Password)
		require.Equal(t, "db.example.com", cfg.TLSConfig.ServerName)
	}
}
