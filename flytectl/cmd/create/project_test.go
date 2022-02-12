package create

import (
	"errors"
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/clierrors"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/project"

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
	project.DefaultProjectConfig.ID = ""
	project.DefaultProjectConfig.Name = ""
	project.DefaultProjectConfig.Labels = map[string]string{}
	project.DefaultProjectConfig.Description = ""
}
func TestCreateProjectFunc(t *testing.T) {
	setup()
	createProjectSetup()
	defer tearDownAndVerify(t, "project Created successfully")
	project.DefaultProjectConfig.ID = projectValue
	project.DefaultProjectConfig.Name = projectValue
	project.DefaultProjectConfig.Labels = map[string]string{}
	project.DefaultProjectConfig.Description = ""
	mockClient.OnRegisterProjectMatch(ctx, projectRegisterRequest).Return(nil, nil)
	err := createProjectsCommand(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "RegisterProject", ctx, projectRegisterRequest)
}

func TestEmptyProjectID(t *testing.T) {
	setup()
	createProjectSetup()
	defer tearDownAndVerify(t, "")
	project.DefaultProjectConfig = &project.ConfigProject{}
	mockClient.OnRegisterProjectMatch(ctx, projectRegisterRequest).Return(nil, nil)
	err := createProjectsCommand(ctx, args, cmdCtx)
	assert.Equal(t, errors.New(clierrors.ErrProjectNotPassed), err)
	mockClient.AssertNotCalled(t, "RegisterProject", ctx, mock.Anything)
}

func TestEmptyProjectName(t *testing.T) {
	setup()
	createProjectSetup()
	defer tearDownAndVerify(t, "")
	project.DefaultProjectConfig.ID = projectValue
	project.DefaultProjectConfig.Labels = map[string]string{}
	project.DefaultProjectConfig.Description = ""
	mockClient.OnRegisterProjectMatch(ctx, projectRegisterRequest).Return(nil, nil)
	err := createProjectsCommand(ctx, args, cmdCtx)
	assert.Equal(t, fmt.Errorf("project name is required flag"), err)
	mockClient.AssertNotCalled(t, "RegisterProject", ctx, mock.Anything)
}
