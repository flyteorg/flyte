package create

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"testing"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

const projectValue = "dummyProject"

var (
	reader                 *os.File
	writer                 *os.File
	err                    error
	ctx                    context.Context
	mockClient             *mocks.AdminServiceClient
	mockOutStream          io.Writer
	args                   []string
	cmdCtx                 cmdCore.CommandContext
	projectRegisterRequest *admin.ProjectRegisterRequest
	stdOut                 *os.File
	stderr                 *os.File
)

func setup() {
	reader, writer, err = os.Pipe()
	if err != nil {
		panic(err)
	}
	stdOut = os.Stdout
	stderr = os.Stderr
	os.Stdout = writer
	os.Stderr = writer
	log.SetOutput(writer)
	mockClient = new(mocks.AdminServiceClient)
	mockOutStream = writer
	cmdCtx = cmdCore.NewCommandContext(mockClient, mockOutStream)
	projectRegisterRequest = &admin.ProjectRegisterRequest{
		Project: &admin.Project{
			Id:          projectValue,
			Name:        projectValue,
			Description: "",
			Labels: &admin.Labels{
				Values: map[string]string{},
			},
		},
	}
}

func teardownAndVerify(t *testing.T, expectedLog string) {
	writer.Close()
	os.Stdout = stdOut
	os.Stderr = stderr
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		assert.Equal(t, expectedLog, buf.String())
	}
}

func TestCreateProjectFunc(t *testing.T) {
	setup()
	defer teardownAndVerify(t, "project Created successfully")
	projectConfig.ID = projectValue
	projectConfig.Name = projectValue
	projectConfig.Labels = map[string]string{}
	projectConfig.Description = ""
	mockClient.OnRegisterProjectMatch(ctx, projectRegisterRequest).Return(nil, nil)
	err := createProjectsCommand(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "RegisterProject", ctx, projectRegisterRequest)
}

func TestEmptyProjectID(t *testing.T) {
	setup()
	defer teardownAndVerify(t, "project ID is required flag")
	projectConfig.Name = projectValue
	projectConfig.Labels = map[string]string{}
	mockClient.OnRegisterProjectMatch(ctx, projectRegisterRequest).Return(nil, nil)
	err := createProjectsCommand(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "RegisterProject", ctx, projectRegisterRequest)
}

func TestEmptyProjectName(t *testing.T) {
	setup()
	defer teardownAndVerify(t, "project ID is required flag")
	projectConfig.ID = projectValue
	projectConfig.Labels = map[string]string{}
	projectConfig.Description = ""
	mockClient.OnRegisterProjectMatch(ctx, projectRegisterRequest).Return(nil, nil)
	err := createProjectsCommand(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "RegisterProject", ctx, projectRegisterRequest)
}
