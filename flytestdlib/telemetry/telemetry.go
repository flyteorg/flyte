// Package telemetry holds the single opt-out for everything Flyte sends to
// third parties. DO_NOT_TRACK=1 (https://consoledonottrack.com) turns it off in
// the SDK, the backend and the console alike; a backend with it set also passes
// it on to the task and app pods it creates.
package telemetry

import (
	"os"
	"strings"
)

// DisableEnvVar is the canonical opt-out, shared with flyte-sdk and the console.
const DisableEnvVar = "DO_NOT_TRACK"

// disableEnvVars all opt out. FLYTE_DISABLE_TELEMETRY opts out of Flyte alone;
// FLYTE_DISABLE_SENTRY is legacy.
var disableEnvVars = []string{DisableEnvVar, "FLYTE_DISABLE_TELEMETRY", "FLYTE_DISABLE_SENTRY"}

// Disabled reports whether any opt-out is set to true, 1 or yes. Unset or
// anything else means enabled.
func Disabled() bool {
	for _, name := range disableEnvVars {
		switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
		case "true", "1", "yes":
			return true
		}
	}
	return false
}
