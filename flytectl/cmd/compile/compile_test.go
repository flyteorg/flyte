package compile

import (
	"context"
	"testing"

	config "github.com/flyteorg/flytectl/cmd/config/subcommand/compile"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	u "github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCompileCommand(t *testing.T) {
	rootCmd := &cobra.Command{
		Long:              "Flytectl is a CLI tool written in Go to interact with the FlyteAdmin service.",
		Short:             "Flytectl CLI tool",
		Use:               "flytectl",
		DisableAutoGenTag: true,
	}
	compileCommand := CreateCompileCommand()
	cmdCore.AddCommands(rootCmd, compileCommand)
	cmdNouns := rootCmd.Commands()
	assert.Equal(t, cmdNouns[0].Use, "compile")
	assert.Equal(t, cmdNouns[0].Flags().Lookup("file").Name, "file")
	// check shorthand
	assert.Equal(t, cmdNouns[0].Short, compileShort)

	// compiling via cobra command
	compileCfg := config.DefaultCompileConfig
	compileCfg.File = "testdata/valid-package.tgz"
	var setup = u.Setup
	s := setup()
	compileCmd := CreateCompileCommand()["compile"]
	err := compileCmd.CmdFunc(context.Background(), []string{}, s.CmdCtx)
	assert.Nil(t, err, "compiling via cmd returns err")

	// calling command with empty file flag
	compileCfg = config.DefaultCompileConfig
	compileCfg.File = ""
	err = compileCmd.CmdFunc(context.Background(), []string{}, s.CmdCtx)
	assert.NotNil(t, err, "calling compile with Empty file flag does not error")
}

func TestCompilePackage(t *testing.T) {
	// valid package contains two workflows
	// with three tasks
	err := compileFromPackage("testdata/valid-package.tgz")
	assert.Nil(t, err, "unable to compile a valid package")

	// invalid gzip header
	err = compileFromPackage("testdata/invalid.tgz")
	assert.NotNil(t, err, "compiling an invalid package returns no error")

	// invalid workflow, types do not match
	err = compileFromPackage("testdata/bad-workflow-package.tgz")
	assert.NotNil(t, err, "compilin an invalid workflow returns no error")

	// testing badly serialized task
	err = compileFromPackage("testdata/invalidtask.tgz")
	assert.NotNil(t, err, "unable to handle invalid task")

	// testing badly serialized launchplan
	err = compileFromPackage("testdata/invalidlaunchplan.tgz")
	assert.NotNil(t, err, "unable to handle invalid launchplan")

	// testing badly serialized workflow
	err = compileFromPackage("testdata/invalidworkflow.tgz")
	assert.NotNil(t, err, "unable to handle invalid workflow")

}
