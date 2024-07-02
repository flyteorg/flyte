package version

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"testing"

	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flytectl/cmd/testutils"
	admin2 "github.com/flyteorg/flyte/flyteidl/clients/go/admin"
	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	stdlibversion "github.com/flyteorg/flyte/flytestdlib/version"
	"github.com/spf13/cobra"
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
		Long:              "Flytectl is a CLI tool written in Go to interact with the FlyteAdmin service.",
		Short:             "Flytectl CLI tool",
		Use:               "flytectl",
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
	s := testutils.Setup()
	defer s.TearDown()
	stdlibversion.Build = ""
	stdlibversion.BuildTime = ""
	stdlibversion.Version = testVersion
	s.MockClient.AdminClient().(*mocks.AdminServiceClient).OnGetVersionMatch(ctx, versionRequest).Return(versionResponse, nil)
	err := getVersion(s.Ctx, []string{}, s.CmdCtx)
	assert.Nil(t, err)
	s.MockClient.AdminClient().(*mocks.AdminServiceClient).AssertCalled(t, "GetVersion", ctx, versionRequest)
}

func TestVersionCommandFuncError(t *testing.T) {
	ctx := context.Background()
	s := testutils.Setup()
	defer s.TearDown()
	stdlibversion.Build = ""
	stdlibversion.BuildTime = ""
	stdlibversion.Version = "v"
	s.MockClient.AdminClient().(*mocks.AdminServiceClient).OnGetVersionMatch(ctx, versionRequest).Return(versionResponse, nil)
	err := getVersion(s.Ctx, []string{}, s.CmdCtx)
	assert.Nil(t, err)
	s.MockClient.AdminClient().(*mocks.AdminServiceClient).AssertCalled(t, "GetVersion", ctx, versionRequest)
}

func TestVersionCommandFuncErr(t *testing.T) {
	ctx := context.Background()
	s := testutils.Setup()
	defer s.TearDown()
	stdlibversion.Build = ""
	stdlibversion.BuildTime = ""
	stdlibversion.Version = testVersion
	s.MockAdminClient.OnGetVersionMatch(ctx, versionRequest).Return(versionResponse, errors.New("error"))
	err := getVersion(s.Ctx, []string{}, s.CmdCtx)
	assert.Nil(t, err)
	s.MockAdminClient.AssertCalled(t, "GetVersion", ctx, versionRequest)
}

func TestVersionUtilFunc(t *testing.T) {
	stdlibversion.Build = ""
	stdlibversion.BuildTime = ""
	stdlibversion.Version = testVersion
	t.Run("Error in getting control plan version", func(t *testing.T) {
		ctx := context.Background()
		mockClient := admin2.InitializeMockClientset()
		adminClient := mockClient.AdminClient().(*mocks.AdminServiceClient)
		mockOutStream := new(io.Writer)
		cmdCtx := cmdCore.NewCommandContext(mockClient, *mockOutStream)
		adminClient.OnGetVersionMatch(ctx, &admin.GetVersionRequest{}).Return(nil, fmt.Errorf("error"))
		err := getControlPlaneVersion(ctx, cmdCtx)
		assert.NotNil(t, err)
	})
	t.Run("Failed in getting version", func(t *testing.T) {
		ctx := context.Background()
		mockClient := admin2.InitializeMockClientset()
		adminClient := mockClient.AdminClient().(*mocks.AdminServiceClient)
		mockOutStream := new(io.Writer)
		cmdCtx := cmdCore.NewCommandContext(mockClient, *mockOutStream)
		adminClient.OnGetVersionMatch(ctx, &admin.GetVersionRequest{}).Return(nil, fmt.Errorf("error"))
		err := getVersion(ctx, []string{}, cmdCtx)
		assert.Nil(t, err)
	})
	t.Run("ClientSet is empty", func(t *testing.T) {
		ctx := context.Background()
		cmdCtx := cmdCore.CommandContext{}
		err := getVersion(ctx, []string{}, cmdCtx)
		assert.Nil(t, err)
	})
}
