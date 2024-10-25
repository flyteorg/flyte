package main

import (
	"github.com/golang/glog"

	"github.com/flyteorg/flyte/cacheservice/cmd/entrypoints"
	"github.com/flyteorg/flyte/flyteadmin/plugins"
)

func main() {
	glog.V(2).Info("Beginning Cache Service")
	err := entrypoints.Execute(plugins.NewRegistry())
	if err != nil {
		panic(err)
	}
}
