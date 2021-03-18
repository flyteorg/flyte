package main

import (
	"fmt"

	"github.com/flyteorg/flytectl/cmd"
)

func main() {
	if err := cmd.ExecuteCmd(); err != nil {
		fmt.Printf("error: %v", err)
	}
}
