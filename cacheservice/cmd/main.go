package main

import (
	"github.com/golang/glog"

	"github.com/flyteorg/flyte/cacheservice/cmd/entrypoints"
)

func main() {
	glog.V(2).Info("Beginning Cache Service")
	err := entrypoints.Execute()
	if err != nil {
		panic(err)
	}
}
