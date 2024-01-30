package main

import (
	"github.com/golang/glog"

	"github.com/flyteorg/flyte/flyteadmin/cmd/scheduler/entrypoints"
)

func main() {
	glog.V(2).Info("Beginning Flyte Scheduler")
	err := entrypoints.Execute()
	if err != nil {
		panic(err)
	}
}
