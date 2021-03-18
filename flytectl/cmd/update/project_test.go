package update

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"os"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

const projectValue = "dummyProject"

var (
	reader               *os.File
	writer               *os.File
	err                  error
	ctx                  context.Context
	mockClient           *mocks.AdminServiceClient
	mockOutStream        io.Writer
	args                 []string
	cmdCtx               cmdCore.CommandContext
	projectUpdateRequest *admin.Project
	stdOut               *os.File
	stderr               *os.File
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
	config.GetConfig().Project = projectValue
	mockClient = new(mocks.AdminServiceClient)
	mockOutStream = writer
	cmdCtx = cmdCore.NewCommandContext(mockClient, mockOutStream)
	projectUpdateRequest = &admin.Project{
		Id:    projectValue,
		State: admin.Project_ACTIVE,
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

func modifyProjectFlags(archiveProject *bool, newArchiveVal bool, activateProject *bool, newActivateVal bool) {
	*archiveProject = newArchiveVal
	*activateProject = newActivateVal
}

func TestActivateProjectFunc(t *testing.T) {
	setup()
	defer teardownAndVerify(t, "Project dummyProject updated to ACTIVE state\n")
	modifyProjectFlags(&(projectConfig.ArchiveProject), false, &(projectConfig.ActivateProject), true)
	mockClient.OnUpdateProjectMatch(ctx, projectUpdateRequest).Return(nil, nil)
	err := updateProjectsFunc(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "UpdateProject", ctx, projectUpdateRequest)
}

func TestActivateProjectFuncWithError(t *testing.T) {
	setup()
	defer teardownAndVerify(t, "Project dummyProject failed to get updated to ACTIVE state due to Error Updating Project\n")
	modifyProjectFlags(&(projectConfig.ArchiveProject), false, &(projectConfig.ActivateProject), true)
	mockClient.OnUpdateProjectMatch(ctx, projectUpdateRequest).Return(nil, errors.New("Error Updating Project"))
	err := updateProjectsFunc(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "UpdateProject", ctx, projectUpdateRequest)
}

func TestArchiveProjectFunc(t *testing.T) {
	setup()
	defer teardownAndVerify(t, "Project dummyProject updated to ARCHIVED state\n")
	modifyProjectFlags(&(projectConfig.ArchiveProject), true, &(projectConfig.ActivateProject), false)
	projectUpdateRequest := &admin.Project{
		Id:    projectValue,
		State: admin.Project_ARCHIVED,
	}
	mockClient.OnUpdateProjectMatch(ctx, projectUpdateRequest).Return(nil, nil)
	err := updateProjectsFunc(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "UpdateProject", ctx, projectUpdateRequest)
}

func TestArchiveProjectFuncWithError(t *testing.T) {
	setup()
	defer teardownAndVerify(t, "Project dummyProject failed to get updated to ARCHIVED state due to Error Updating Project\n")
	modifyProjectFlags(&(projectConfig.ArchiveProject), true, &(projectConfig.ActivateProject), false)
	projectUpdateRequest := &admin.Project{
		Id:    projectValue,
		State: admin.Project_ARCHIVED,
	}
	mockClient.OnUpdateProjectMatch(ctx, projectUpdateRequest).Return(nil, errors.New("Error Updating Project"))
	err := updateProjectsFunc(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "UpdateProject", ctx, projectUpdateRequest)
}

func TestEmptyProjectInput(t *testing.T) {
	setup()
	defer teardownAndVerify(t, "Project  not found\n")
	config.GetConfig().Project = ""
	modifyProjectFlags(&(projectConfig.ArchiveProject), false, &(projectConfig.ActivateProject), true)
	mockClient.OnUpdateProjectMatch(ctx, projectUpdateRequest).Return(nil, nil)
	err := updateProjectsFunc(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertNotCalled(t, "UpdateProject", ctx, projectUpdateRequest)
}

func TestInvalidInput(t *testing.T) {
	setup()
	defer teardownAndVerify(t, "Invalid state passed. Specify either activate or archive\n")
	modifyProjectFlags(&(projectConfig.ArchiveProject), false, &(projectConfig.ActivateProject), false)
	mockClient.OnUpdateProjectMatch(ctx, projectUpdateRequest).Return(nil, nil)
	err := updateProjectsFunc(ctx, args, cmdCtx)
	assert.NotNil(t, err)
	mockClient.AssertNotCalled(t, "UpdateProject", ctx, projectUpdateRequest)
}
