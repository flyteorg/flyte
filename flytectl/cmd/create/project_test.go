package create

import (
	"errors"
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/clierrors"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/project"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const projectValue = "dummyProject"

var (
	projectRegisterRequest *admin.ProjectRegisterRequest
)

func createProjectSetup() {
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
	config.GetConfig().Project = ""
}
func TestCreateProjectFunc(t *testing.T) {
	s := setup()
	createProjectSetup()
	defer s.TearDownAndVerify(t, "project created successfully.")
	project.DefaultProjectConfig.ID = projectValue
	project.DefaultProjectConfig.Name = projectValue
	project.DefaultProjectConfig.Labels = map[string]string{}
	project.DefaultProjectConfig.Description = ""
	s.MockAdminClient.OnRegisterProjectMatch(s.Ctx, projectRegisterRequest).Return(nil, nil)
	err := createProjectsCommand(s.Ctx, []string{}, s.CmdCtx)
	assert.Nil(t, err)
	s.MockAdminClient.AssertCalled(t, "RegisterProject", s.Ctx, projectRegisterRequest)
}

func TestEmptyProjectID(t *testing.T) {
	s := setup()
	createProjectSetup()
	defer s.TearDownAndVerify(t, "")
	project.DefaultProjectConfig = &project.ConfigProject{}
	s.MockAdminClient.OnRegisterProjectMatch(s.Ctx, projectRegisterRequest).Return(nil, nil)
	err := createProjectsCommand(s.Ctx, []string{}, s.CmdCtx)
	assert.Equal(t, errors.New(clierrors.ErrProjectNotPassed), err)
	s.MockAdminClient.AssertNotCalled(t, "RegisterProject", s.Ctx, mock.Anything)
}

func TestEmptyProjectName(t *testing.T) {
	s := setup()
	createProjectSetup()
	defer s.TearDownAndVerify(t, "")
	project.DefaultProjectConfig.ID = projectValue
	project.DefaultProjectConfig.Labels = map[string]string{}
	project.DefaultProjectConfig.Description = ""
	s.MockAdminClient.OnRegisterProjectMatch(s.Ctx, projectRegisterRequest).Return(nil, nil)
	err := createProjectsCommand(s.Ctx, []string{}, s.CmdCtx)
	assert.Equal(t, fmt.Errorf("project name is a required flag"), err)
	s.MockAdminClient.AssertNotCalled(t, "RegisterProject", s.Ctx, mock.Anything)
}
