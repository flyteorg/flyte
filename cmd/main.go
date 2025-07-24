package main

import (
	"github.com/golang/glog"

	"github.com/flyteorg/flyte/cmd/single"
	_ "github.com/flyteorg/flyte/flytepropeller/plugins"
)

func main() {
	glog.V(2).Info("Starting Flyte")
	err := single.Execute()
	if err != nil {
		panic(err)
	}
}
