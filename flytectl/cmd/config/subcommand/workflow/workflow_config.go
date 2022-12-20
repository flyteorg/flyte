package workflow

import (
	"github.com/flyteorg/flytectl/pkg/filters"
)

//go:generate pflags Config --default-var DefaultConfig --bind-default-var

var (
	wfDefaultFilter = filters.Filters{
		Limit: filters.DefaultLimit,
		Page:  1,
	}
	DefaultConfig = &Config{
		Filter: wfDefaultFilter,
	}
)

// Config commandline configuration
type Config struct {
	Version string          `json:"version" pflag:",version of the workflow to be fetched."`
	Latest  bool            `json:"latest" pflag:", flag to indicate to fetch the latest version, version flag will be ignored in this case"`
	Filter  filters.Filters `json:"filter" pflag:","`
}
