package task

import "github.com/flyteorg/flytectl/pkg/filters"

//go:generate pflags Config --default-var DefaultConfig --bind-default-var
var (
	DefaultConfig = &Config{
		Filter: filters.DefaultFilter,
	}
)

// Config
type Config struct {
	ExecFile string          `json:"execFile" pflag:",execution file name to be used for generating execution spec of a single task."`
	Version  string          `json:"version" pflag:",version of the task to be fetched."`
	Latest   bool            `json:"latest" pflag:", flag to indicate to fetch the latest version, version flag will be ignored in this case"`
	Filter   filters.Filters `json:"filter" pflag:","`
}
