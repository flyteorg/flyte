package main

import (
	"github.com/flyteorg/flyteadmin/cmd/entrypoints"
	"github.com/golang/glog"
)

func main() {
	glog.V(2).Info("Beginning Flyte Controller")
	err := entrypoints.Execute()
	if err != nil {
		panic(err)
	}
}
