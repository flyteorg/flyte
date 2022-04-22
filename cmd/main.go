package main

import (
	"github.com/flyteorg/flyte/cmd/single"
	_ "github.com/flyteorg/flytepropeller/plugins"
	"github.com/golang/glog"
)

func main() {
	glog.V(2).Info("Starting Flyte")
	err := single.Execute()
	if err != nil {
		panic(err)
	}
}
