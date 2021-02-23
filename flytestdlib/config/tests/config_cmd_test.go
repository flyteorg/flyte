package tests

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/flyteorg/flytestdlib/config"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOutput(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

func TestDiscoverCommand(t *testing.T) {
	for _, provider := range providers {
		t.Run(fmt.Sprintf(testNameFormatter, provider(config.Options{}).ID(), "No config file"), func(t *testing.T) {
			cmd := config.NewConfigCommand(provider)
			output, err := executeCommand(cmd, config.CommandDiscover)
			assert.NoError(t, err)
			assert.Contains(t, output, "Couldn't find a config file.")
		})

		t.Run(fmt.Sprintf(testNameFormatter, provider(config.Options{}).ID(), "Valid config file"), func(t *testing.T) {
			dir, err := os.Getwd()
			assert.NoError(t, err)
			wd := os.ExpandEnv("$PWD/testdata")
			err = os.Chdir(wd)
			assert.NoError(t, err)
			defer func() { assert.NoError(t, os.Chdir(dir)) }()

			cmd := config.NewConfigCommand(provider)
			output, err := executeCommand(cmd, config.CommandDiscover)
			assert.NoError(t, err)
			assert.Contains(t, output, "Config")
		})
	}
}

func TestValidateCommand(t *testing.T) {
	for _, provider := range providers {
		t.Run(fmt.Sprintf(testNameFormatter, provider(config.Options{}).ID(), "No config file"), func(t *testing.T) {
			cmd := config.NewConfigCommand(provider)
			output, err := executeCommand(cmd, config.CommandValidate)
			assert.NoError(t, err)
			assert.Contains(t, output, "Couldn't find a config file.")
		})

		t.Run(fmt.Sprintf(testNameFormatter, provider(config.Options{}).ID(), "Invalid Config file"), func(t *testing.T) {
			dir, err := os.Getwd()
			assert.NoError(t, err)
			wd := os.ExpandEnv("$PWD/testdata")
			err = os.Chdir(wd)
			assert.NoError(t, err)
			defer func() { assert.NoError(t, os.Chdir(dir)) }()

			cmd := config.NewConfigCommand(provider)
			output, err := executeCommand(cmd, config.CommandValidate, "--file=bad_config.yaml", "--strict")
			assert.Error(t, err)
			assert.Contains(t, output, "Failed")
		})

		t.Run(fmt.Sprintf(testNameFormatter, provider(config.Options{}).ID(), "Valid config file"), func(t *testing.T) {
			dir, err := os.Getwd()
			assert.NoError(t, err)
			wd := os.ExpandEnv("$PWD/testdata")
			err = os.Chdir(wd)
			assert.NoError(t, err)
			defer func() { assert.NoError(t, os.Chdir(dir)) }()

			cmd := config.NewConfigCommand(provider)
			output, err := executeCommand(cmd, config.CommandValidate)
			assert.NoError(t, err)
			assert.Contains(t, output, "successfully")
		})
	}
}
