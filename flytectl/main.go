package main

import "github.com/lyft/flytectl/cmd"

func main() {
	if err := cmd.ExecuteCmd(); err != nil {
		panic(err)
	}
}
