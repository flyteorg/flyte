package main

import (
	"github.com/flyteorg/flyte/cmd/single"
	"github.com/golang/glog"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/k8s"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/pod"
)

func main() {
	glog.V(2).Info("Starting Flyte")
	err := single.Execute()
	if err != nil {
		panic(err)
	}
}
