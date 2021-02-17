package main

import (
	"fmt"

	"github.com/lyft/flytectl/cmd"
)

func main() {
	if err := cmd.ExecuteCmd(); err != nil {
		fmt.Printf("error: %v", err)
	}
}
