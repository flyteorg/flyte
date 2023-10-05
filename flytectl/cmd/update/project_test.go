package update

import (
	"errors"
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/project"

	"github.com/flyteorg/flytectl/clierrors"
	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
)

const projectValue = "dummyProject"

var (
	projectUpdateRequest *admin.Project
)

func updateProjectSetup() {
	projectUpdateRequest = &admin.Project{
		Id:    projectValue,
		State: admin.Project_ACTIVE,
	}
}

func modifyProjectFlags(newArchiveVal bool, newActivateVal bool) {
	project.DefaultProjectConfig.ArchiveProject = newArchiveVal
	project.DefaultProjectConfig.Archive = newArchiveVal
	project.DefaultProjectConfig.ActivateProject = newActivateVal
	project.DefaultProjectConfig.Activate = newActivateVal
}

func TestActivateProjectFunc(t *testing.T) {
	s := setup()
	updateProjectSetup()
	config.GetConfig().Project = projectValue
	project.DefaultProjectConfig.Name = projectValue
	modifyProjectFlags(false, true)
	projectUpdateRequest = &admin.Project{
		Id:   projectValue,
		Name: projectValue,
		Labels: &admin.Labels{
			Values: map[string]string{},
		},
		State: admin.Project_ACTIVE,
	}
	s.MockAdminClient.OnUpdateProjectMatch(s.Ctx, projectUpdateRequest).Return(nil, nil)
	err := updateProjectsFunc(s.Ctx, []string{}, s.CmdCtx)
	assert.Nil(t, err)
	s.MockAdminClient.AssertCalled(t, "UpdateProject", s.Ctx, projectUpdateRequest)
	s.TearDownAndVerify(t, "Project dummyProject updated\n")
}

func TestActivateProjectFuncWithError(t *testing.T) {
	s := setup()
	updateProjectSetup()
	config.GetConfig().Project = projectValue
	project.DefaultProjectConfig.Name = projectValue
	modifyProjectFlags(false, true)
	projectUpdateRequest = &admin.Project{
		Id:   projectValue,
		Name: projectValue,
		Labels: &admin.Labels{
			Values: map[string]string{},
		},
		State: admin.Project_ACTIVE,
	}
	s.MockAdminClient.OnUpdateProjectMatch(s.Ctx, projectUpdateRequest).Return(nil, errors.New("Error Updating Project"))
	err := updateProjectsFunc(s.Ctx, []string{}, s.CmdCtx)
	assert.NotNil(t, err)
	s.MockAdminClient.AssertCalled(t, "UpdateProject", s.Ctx, projectUpdateRequest)
	s.TearDownAndVerify(t, "Project dummyProject failed to update due to Error Updating Project\n")
}

func TestArchiveProjectFunc(t *testing.T) {
	s := setup()
	updateProjectSetup()
	config.GetConfig().Project = projectValue
	project.DefaultProjectConfig = &project.ConfigProject{}
	project.DefaultProjectConfig.Name = projectValue
	modifyProjectFlags(true, false)
	projectUpdateRequest = &admin.Project{
		Id:   projectValue,
		Name: projectValue,
		Labels: &admin.Labels{
			Values: nil,
		},
		State: admin.Project_ARCHIVED,
	}
	s.MockAdminClient.OnUpdateProjectMatch(s.Ctx, projectUpdateRequest).Return(nil, nil)
	err := updateProjectsFunc(s.Ctx, []string{}, s.CmdCtx)
	assert.Nil(t, err)
	s.MockAdminClient.AssertCalled(t, "UpdateProject", s.Ctx, projectUpdateRequest)
	s.TearDownAndVerify(t, "Project dummyProject updated\n")
}

func TestArchiveProjectFuncWithError(t *testing.T) {
	s := setup()
	updateProjectSetup()
	project.DefaultProjectConfig.Name = projectValue
	project.DefaultProjectConfig.Labels = map[string]string{}
	modifyProjectFlags(true, false)
	projectUpdateRequest = &admin.Project{
		Id:   projectValue,
		Name: projectValue,
		Labels: &admin.Labels{
			Values: map[string]string{},
		},
		State: admin.Project_ARCHIVED,
	}
	s.MockAdminClient.OnUpdateProjectMatch(s.Ctx, projectUpdateRequest).Return(nil, errors.New("Error Updating Project"))
	err := updateProjectsFunc(s.Ctx, []string{}, s.CmdCtx)
	assert.NotNil(t, err)
	s.MockAdminClient.AssertCalled(t, "UpdateProject", s.Ctx, projectUpdateRequest)
	s.TearDownAndVerify(t, "Project dummyProject failed to update"+
		" due to Error Updating Project\n")
}

func TestEmptyProjectInput(t *testing.T) {
	s := setup()
	updateProjectSetup()
	config.GetConfig().Project = ""
	modifyProjectFlags(false, true)
	err := updateProjectsFunc(s.Ctx, []string{}, s.CmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf(clierrors.ErrProjectNotPassed), err)
}

func TestInvalidInput(t *testing.T) {
	s := setup()
	updateProjectSetup()
	config.GetConfig().Project = projectValue
	project.DefaultProjectConfig.Name = projectValue
	modifyProjectFlags(true, true)
	err := updateProjectsFunc(s.Ctx, []string{}, s.CmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf(clierrors.ErrInvalidStateUpdate), err)
	s.TearDownAndVerify(t, "")
}
