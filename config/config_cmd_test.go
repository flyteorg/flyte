package config

import (
	"bytes"
	"context"
	"flag"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

type MockAccessor struct {
}

func (MockAccessor) ID() string {
	panic("implement me")
}

func (MockAccessor) InitializeFlags(cmdFlags *flag.FlagSet) {
}

func (MockAccessor) InitializePflags(cmdFlags *pflag.FlagSet) {
}

func (MockAccessor) UpdateConfig(ctx context.Context) error {
	return nil
}

func (MockAccessor) ConfigFilesUsed() []string {
	return []string{"test"}
}

func (MockAccessor) RefreshFromConfig() error {
	return nil
}

func newMockAccessor(options Options) Accessor {
	return MockAccessor{}
}

func executeCommandC(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs(args)

	_, err = root.ExecuteC()

	return buf.String(), err
}

func TestNewConfigCommand(t *testing.T) {
	cmd := NewConfigCommand(newMockAccessor)
	assert.NotNil(t, cmd)

	output, err := executeCommandC(cmd, CommandDiscover)
	assert.NoError(t, err)
	assert.Contains(t, output, "test")

	output, err = executeCommandC(cmd, CommandValidate)
	assert.NoError(t, err)
	assert.Contains(t, output, "test")
}
