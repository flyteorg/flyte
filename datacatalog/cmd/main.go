package main

import (
	"github.com/golang/glog"
	"github.com/lyft/datacatalog/cmd/entrypoints"
)

func main() {
	glog.V(2).Info("Beginning Data Catalog")
	err := entrypoints.Execute()
	if err != nil {
		panic(err)
	}
}
