package execution

import (
	"github.com/flyteorg/flytectl/pkg/filters"
)

//go:generate pflags Config --default-var DefaultConfig --bind-default-var
var (
	DefaultConfig = &Config{
		Filter: filters.DefaultFilter,
	}
)

// Config stores the flags required by get execution
type Config struct {
	Filter  filters.Filters `json:"filter" pflag:","`
	Details bool            `json:"details" pflag:",gets node execution details. Only applicable for single execution name i.e get execution name --details"`
	NodeID  string          `json:"nodeId" pflag:",get task executions for given node name."`
}
