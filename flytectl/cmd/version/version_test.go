package version

import (
	"context"
	"fmt"
	"io"
	"testing"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

func TestListExecutionFunc(t *testing.T) {
	ctx := context.Background()
	var args []string
	mockClient := new(mocks.AdminServiceClient)
	mockOutStream := new(io.Writer)
	cmdCtx := cmdCore.NewCommandContext(mockClient, *mockOutStream)
	versionRequest := &admin.GetVersionRequest{}
	versionResponse := &admin.GetVersionResponse{
		ControlPlaneVersion: &admin.Version{
			Build:     "",
			BuildTime: "",
			Version:   "",
		},
	}
	mockClient.OnGetVersionMatch(ctx, versionRequest).Return(versionResponse, nil)
	err := getVersion(ctx, args, cmdCtx)
	fmt.Println(err)
	assert.Nil(t, nil)
	mockClient.AssertCalled(t, "GetVersion", ctx, versionRequest)
}
