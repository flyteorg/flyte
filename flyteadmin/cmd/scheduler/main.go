package main

import (
	entrypoints2 "github.com/flyteorg/flyteadmin/cmd/scheduler/entrypoints"
	"github.com/golang/glog"
)

func main() {
	glog.V(2).Info("Beginning Flyte Scheduler")
	err := entrypoints2.Execute()
	if err != nil {
		panic(err)
	}
}
