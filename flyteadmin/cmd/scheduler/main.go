package main

import (
	"github.com/flyteorg/flyte/flyteadmin/cmd/scheduler/entrypoints"
	"github.com/golang/glog"
)

func main() {
	glog.V(2).Info("Beginning Flyte Scheduler")
	err := entrypoints.Execute()
	if err != nil {
		panic(err)
	}
}
