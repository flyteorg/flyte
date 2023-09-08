package main

import (
	"github.com/flyteorg/flyte/datacatalog/cmd/entrypoints"
	"github.com/golang/glog"
)

func main() {
	glog.V(2).Info("Beginning Data Catalog")
	err := entrypoints.Execute()
	if err != nil {
		panic(err)
	}
}
