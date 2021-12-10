package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCompletionCmdIntegration(t *testing.T) {
	rootCmd := &cobra.Command{
		Long:              "FlyteCTL is CLI tool written in go to interact with Flyteadmin service",
		Short:             "FlyteCTL CLI tool",
		Use:               "FlyteCTL",
		DisableAutoGenTag: true,
	}

	err := completionCmd.RunE(rootCmd, []string{"bash"})
	assert.Nil(t, err)
	err = completionCmd.RunE(rootCmd, []string{"zsh"})
	assert.Nil(t, err)
	err = completionCmd.RunE(rootCmd, []string{"fish"})
	assert.Nil(t, err)
	err = completionCmd.RunE(rootCmd, []string{"powershell"})
	assert.Nil(t, err)
}
