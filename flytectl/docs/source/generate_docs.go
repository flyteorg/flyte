package main

import "github.com/flyteorg/flyte/flytectl/cmd"

func main() {
	if err := cmd.GenerateDocs(); err != nil {
		panic(err)
	}
}
