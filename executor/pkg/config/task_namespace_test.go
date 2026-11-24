package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllowedTaskNamespacesFlag(t *testing.T) {
	flags := (Config{}).GetPFlagSet("executor.")
	name := "executor.allowedTaskNamespaces"
	defaults, err := flags.GetStringSlice(name)
	require.NoError(t, err)
	require.Empty(t, defaults)
	require.NoError(t, flags.Set(name, "gpu-workers,cpu-workers"))
	values, err := flags.GetStringSlice(name)
	require.NoError(t, err)
	require.Equal(t, []string{"gpu-workers", "cpu-workers"}, values)
}
