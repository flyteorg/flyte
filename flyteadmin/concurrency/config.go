package concurrency

import (
	"github.com/flyteorg/flyte/flytestdlib/config"
)

const configSectionKey = "concurrency"

// Config contains settings for the concurrency controller
type Config struct {
	// Enable determines whether to enable the concurrency controller
	Enable bool `json:"enable" pflag:",Enable the execution concurrency controller"`

	// ProcessingInterval defines how often to check for pending executions
	ProcessingInterval config.Duration `json:"processingInterval" pflag:",How often to process pending executions"`

	// Workers defines the number of worker goroutines to process executions
	Workers int `json:"workers" pflag:",Number of worker goroutines for processing executions"`

	// MaxRetries defines the maximum number of retries for processing an execution
	MaxRetries int `json:"maxRetries" pflag:",Maximum number of retries for processing an execution"`

	// LaunchPlanRefreshInterval defines how often to refresh launch plan concurrency information
	LaunchPlanRefreshInterval config.Duration `json:"launchPlanRefreshInterval" pflag:",How often to refresh launch plan concurrency information"`
}
