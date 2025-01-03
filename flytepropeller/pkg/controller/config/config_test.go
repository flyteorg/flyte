package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSupportedSDKVersion(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name           string
		config         LiteralOffloadingConfig
		sdk            string
		version        string
		expectedResult bool
	}{
		{
			name: "supported version",
			config: LiteralOffloadingConfig{
				SupportedSDKVersions: map[string]string{
					"flytekit": "0.16.0",
				},
			},
			sdk:            "flytekit",
			version:        "0.16.0",
			expectedResult: true,
		},
		{
			name: "unsupported version",
			config: LiteralOffloadingConfig{
				SupportedSDKVersions: map[string]string{
					"flytekit": "0.16.0",
				},
			},
			sdk:            "flytekit",
			version:        "0.15.0",
			expectedResult: false,
		},
		{
			name: "unsupported SDK",
			config: LiteralOffloadingConfig{
				SupportedSDKVersions: map[string]string{
					"flytekit": "0.16.0",
				},
			},
			sdk:            "unknown",
			version:        "0.16.0",
			expectedResult: false,
		},
		{
			name: "invalid version",
			config: LiteralOffloadingConfig{
				SupportedSDKVersions: map[string]string{
					"flytekit": "0.16.0",
				},
			},
			sdk:            "flytekit",
			version:        "invalid",
			expectedResult: false,
		},
		{
			name: "invalid constraint",
			config: LiteralOffloadingConfig{
				SupportedSDKVersions: map[string]string{
					"flytekit": "invalid",
				},
			},
			sdk:            "flytekit",
			version:        "0.16.0",
			expectedResult: false,
		},
		{
			name: "supported dev version",
			config: LiteralOffloadingConfig{
				SupportedSDKVersions: map[string]string{
					"flytekit": "1.13.4",
				},
			},
			sdk:            "flytekit",
			version:        "1.13.4.dev12+g990b450ea.d20240917",
			expectedResult: true,
		},
		{
			name: "supported beta version",
			config: LiteralOffloadingConfig{
				SupportedSDKVersions: map[string]string{
					"flytekit": "1.13.4",
				},
			},
			sdk:            "flytekit",
			version:        "v1.13.6b0",
			expectedResult: true,
		},
		{
			name: "unsupported dev version",
			config: LiteralOffloadingConfig{
				SupportedSDKVersions: map[string]string{
					"flytekit": "1.13.4",
				},
			},
			sdk:            "flytekit",
			version:        "1.13.3.dev12+g990b450ea.d20240917",
			expectedResult: false,
		},
		{
			name: "unsupported beta version",
			config: LiteralOffloadingConfig{
				SupportedSDKVersions: map[string]string{
					"flytekit": "1.13.4",
				},
			},
			sdk:            "flytekit",
			version:        "v1.13.3b0",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsSupportedSDKVersion(ctx, tt.sdk, tt.version)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
