package main

import (
	"github.com/golang/glog"
	"github.com/lyft/flyteadmin/cmd/entrypoints"
)

func main() {
	glog.V(2).Info("Beginning Flyte Controller")
	err := entrypoints.Execute()
	if err != nil {
		panic(err)
	}
}
