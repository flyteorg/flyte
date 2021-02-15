package main

import "github.com/lyft/flytectl/cmd"

func main() {
	if err := cmd.GenerateDocs(); err != nil {
		panic(err)
	}
}
