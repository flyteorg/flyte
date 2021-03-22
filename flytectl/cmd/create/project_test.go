package create

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const projectValue = "dummyProject"

var (
	projectRegisterRequest *admin.ProjectRegisterRequest
)

func createProjectSetup() {
	ctx = testutils.Ctx
	cmdCtx = testutils.CmdCtx
	mockClient = testutils.MockClient
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
	projectConfig.ID = ""
	projectConfig.Name = ""
	projectConfig.Labels = map[string]string{}
	projectConfig.Description = ""
}
func TestCreateProjectFunc(t *testing.T) {
	setup()
	createProjectSetup()
	defer tearDownAndVerify(t, "project Created successfully")
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
	createProjectSetup()
	defer tearDownAndVerify(t, "")
	projectConfig.Name = projectValue
	projectConfig.Labels = map[string]string{}
	mockClient.OnRegisterProjectMatch(ctx, projectRegisterRequest).Return(nil, nil)
	err := createProjectsCommand(ctx, args, cmdCtx)
	assert.Equal(t, fmt.Errorf("project ID is required flag"), err)
	mockClient.AssertNotCalled(t, "RegisterProject", ctx, mock.Anything)
}

func TestEmptyProjectName(t *testing.T) {
	setup()
	createProjectSetup()
	defer tearDownAndVerify(t, "")
	projectConfig.ID = projectValue
	projectConfig.Labels = map[string]string{}
	projectConfig.Description = ""
	mockClient.OnRegisterProjectMatch(ctx, projectRegisterRequest).Return(nil, nil)
	err := createProjectsCommand(ctx, args, cmdCtx)
	assert.Equal(t, fmt.Errorf("project name is required flag"), err)
	mockClient.AssertNotCalled(t, "RegisterProject", ctx, mock.Anything)
}
