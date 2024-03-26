package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func requiredFlags(cmd *cobra.Command, flags ...string) error {
	for _, flag := range flags {
		f := cmd.Flag(flag)
		if f == nil {
			return fmt.Errorf("unable to find Key [%v]", flag)
		}
	}

	return nil
}
