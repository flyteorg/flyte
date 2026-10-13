package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisabled(t *testing.T) {
	for _, name := range disableEnvVars {
		for _, value := range []string{"true", "1", "yes", "TRUE", " true "} {
			t.Run(name+"="+value, func(t *testing.T) {
				for _, n := range disableEnvVars {
					t.Setenv(n, "")
				}
				t.Setenv(name, value)
				assert.True(t, Disabled())
			})
		}
	}

	for _, value := range []string{"", "false", "0", "no", "garbage"} {
		t.Run("enabled="+value, func(t *testing.T) {
			for _, n := range disableEnvVars {
				t.Setenv(n, "")
			}
			t.Setenv("FLYTE_DISABLE_TELEMETRY", value)
			assert.False(t, Disabled())
		})
	}
}
