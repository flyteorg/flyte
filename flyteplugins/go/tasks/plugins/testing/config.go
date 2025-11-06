package testing

import (
	"time"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/config"
	flytestdconfig "github.com/flyteorg/flyte/v2/flytestdlib/config"
)

//go:generate pflags Config --default-var=defaultConfig

var (
	defaultConfig = Config{
		SleepDuration: flytestdconfig.Duration{Duration: 0 * time.Second},
	}

	ConfigSection = config.MustRegisterSubSection(echoTaskType, &defaultConfig)
)

type Config struct {
	// SleepDuration indicates the amount of time before transitioning to success
	SleepDuration flytestdconfig.Duration `json:"sleep-duration" pflag:",Indicates the amount of time before transitioning to success"`
}
