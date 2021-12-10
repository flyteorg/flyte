package version

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"testing"

	"github.com/spf13/cobra"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	stdlibversion "github.com/flyteorg/flytestdlib/version"
	"github.com/stretchr/testify/assert"
)

var (
	versionRequest  = &admin.GetVersionRequest{}
	testVersion     = "v0.1.20"
	versionResponse = &admin.GetVersionResponse{
		ControlPlaneVersion: &admin.Version{
			Build:     "",
			BuildTime: "",
			Version:   testVersion,
		},
	}
)

func TestVersionCommand(t *testing.T) {
	rootCmd := &cobra.Command{
		Long:              "FlyteCTL is CLI tool written in go to interact with Flyteadmin service",
		Short:             "FlyteCTL CLI tool",
		Use:               "FlyteCTL",
		DisableAutoGenTag: true,
	}
	versionCommand := GetVersionCommand(rootCmd)
	cmdCore.AddCommands(rootCmd, versionCommand)
	fmt.Println(rootCmd.Commands())
	assert.Equal(t, len(rootCmd.Commands()), 1)
	cmdNouns := rootCmd.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})

	assert.Equal(t, cmdNouns[0].Use, "version")
	assert.Equal(t, cmdNouns[0].Short, versionCmdShort)
	assert.Equal(t, cmdNouns[0].Long, versionCmdLong)
}

func TestVersionCommandFunc(t *testing.T) {
	ctx := context.Background()
	var args []string
	mockClient := new(mocks.AdminServiceClient)
	mockOutStream := new(io.Writer)
	cmdCtx := cmdCore.NewCommandContext(mockClient, *mockOutStream)
	stdlibversion.Build = ""
	stdlibversion.BuildTime = ""
	stdlibversion.Version = testVersion
	mockClient.OnGetVersionMatch(ctx, versionRequest).Return(versionResponse, nil)
	err := getVersion(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "GetVersion", ctx, versionRequest)
}

func TestVersionCommandFuncError(t *testing.T) {
	ctx := context.Background()
	var args []string
	mockClient := new(mocks.AdminServiceClient)
	mockOutStream := new(io.Writer)
	cmdCtx := cmdCore.NewCommandContext(mockClient, *mockOutStream)
	stdlibversion.Build = ""
	stdlibversion.BuildTime = ""
	stdlibversion.Version = "v"
	mockClient.OnGetVersionMatch(ctx, versionRequest).Return(versionResponse, nil)
	err := getVersion(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "GetVersion", ctx, versionRequest)
}

func TestVersionCommandFuncErr(t *testing.T) {
	ctx := context.Background()
	var args []string
	mockClient := new(mocks.AdminServiceClient)
	mockOutStream := new(io.Writer)
	cmdCtx := cmdCore.NewCommandContext(mockClient, *mockOutStream)
	stdlibversion.Build = ""
	stdlibversion.BuildTime = ""
	stdlibversion.Version = testVersion
	mockClient.OnGetVersionMatch(ctx, versionRequest).Return(versionResponse, errors.New("error"))
	err := getVersion(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "GetVersion", ctx, versionRequest)
}

func TestVersionUtilFunc(t *testing.T) {
	stdlibversion.Build = ""
	stdlibversion.BuildTime = ""
	stdlibversion.Version = testVersion
	t.Run("Error in getting control plan version", func(t *testing.T) {
		ctx := context.Background()
		mockClient := new(mocks.AdminServiceClient)
		mockOutStream := new(io.Writer)
		cmdCtx := cmdCore.NewCommandContext(mockClient, *mockOutStream)
		mockClient.OnGetVersionMatch(ctx, &admin.GetVersionRequest{}).Return(nil, fmt.Errorf("error"))
		err := getControlPlaneVersion(ctx, cmdCtx)
		assert.NotNil(t, err)
	})
	t.Run("Failed in getting version", func(t *testing.T) {
		ctx := context.Background()
		mockClient := new(mocks.AdminServiceClient)
		mockOutStream := new(io.Writer)
		cmdCtx := cmdCore.NewCommandContext(mockClient, *mockOutStream)
		mockClient.OnGetVersionMatch(ctx, &admin.GetVersionRequest{}).Return(nil, fmt.Errorf("error"))
		err := getVersion(ctx, []string{}, cmdCtx)
		assert.Nil(t, err)
	})

}
