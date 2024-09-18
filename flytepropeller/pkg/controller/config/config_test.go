package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSupportedSDKVersion(t *testing.T) {
	t.Run("supported version", func(t *testing.T) {
		config := LiteralOffloadingConfig{
			SupportedSDKVersions: map[string]string{
				"flytekit": "0.16.0",
			},
		}
		assert.True(t, config.IsSupportedSDKVersion("flytekit", "0.16.0"))
	})

	t.Run("unsupported version", func(t *testing.T) {
		config := LiteralOffloadingConfig{
			SupportedSDKVersions: map[string]string{
				"flytekit": "0.16.0",
			},
		}
		assert.False(t, config.IsSupportedSDKVersion("flytekit", "0.15.0"))
	})

	t.Run("unsupported SDK", func(t *testing.T) {
		config := LiteralOffloadingConfig{
			SupportedSDKVersions: map[string]string{
				"flytekit": "0.16.0",
			},
		}
		assert.False(t, config.IsSupportedSDKVersion("unknown", "0.16.0"))
	})

	t.Run("invalid version", func(t *testing.T) {
		config := LiteralOffloadingConfig{
			SupportedSDKVersions: map[string]string{
				"flytekit": "0.16.0",
			},
		}
		assert.False(t, config.IsSupportedSDKVersion("flytekit", "invalid"))
	})

	t.Run("invalid constraint", func(t *testing.T) {
		config := LiteralOffloadingConfig{
			SupportedSDKVersions: map[string]string{
				"flytekit": "invalid",
			},
		}
		assert.False(t, config.IsSupportedSDKVersion("flytekit", "0.16.0"))
	})

	t.Run("supported dev version", func(t *testing.T) {
		config := LiteralOffloadingConfig{
			SupportedSDKVersions: map[string]string{
				"flytekit": "1.13.4",
			},
		}
		assert.True(t, config.IsSupportedSDKVersion("flytekit", "1.13.4.dev12+g990b450ea.d20240917"))
	})
	t.Run("supported beta version", func(t *testing.T) {
		config := LiteralOffloadingConfig{
			SupportedSDKVersions: map[string]string{
				"flytekit": "1.13.4",
			},
		}
		assert.True(t, config.IsSupportedSDKVersion("flytekit", "v1.13.6b0"))
	})
}
