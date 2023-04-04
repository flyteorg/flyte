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

func modifyProjectFlags(archiveProject *bool, newArchiveVal bool, activateProject *bool, newActivateVal bool) {
	*archiveProject = newArchiveVal
	*activateProject = newActivateVal
}

func TestActivateProjectFunc(t *testing.T) {
	s := setup()
	updateProjectSetup()
	config.GetConfig().Project = projectValue
	project.DefaultProjectConfig.Name = projectValue
	modifyProjectFlags(&(project.DefaultProjectConfig.ArchiveProject), false, &(project.DefaultProjectConfig.ActivateProject), true)
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
	tearDownAndVerify(t, s.Writer, "Project dummyProject updated\n")
}

func TestActivateProjectFuncWithError(t *testing.T) {
	s := setup()
	updateProjectSetup()
	config.GetConfig().Project = projectValue
	project.DefaultProjectConfig.Name = projectValue
	modifyProjectFlags(&(project.DefaultProjectConfig.ArchiveProject), false, &(project.DefaultProjectConfig.ActivateProject), true)
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
	tearDownAndVerify(t, s.Writer, "Project dummyProject failed to update due to Error Updating Project\n")
}

func TestArchiveProjectFunc(t *testing.T) {
	s := setup()
	updateProjectSetup()
	config.GetConfig().Project = projectValue
	project.DefaultProjectConfig = &project.ConfigProject{}
	project.DefaultProjectConfig.Name = projectValue
	modifyProjectFlags(&(project.DefaultProjectConfig.ArchiveProject), true, &(project.DefaultProjectConfig.ActivateProject), false)
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
	tearDownAndVerify(t, s.Writer, "Project dummyProject updated\n")
}

func TestArchiveProjectFuncWithError(t *testing.T) {
	s := setup()
	updateProjectSetup()
	project.DefaultProjectConfig.Name = projectValue
	project.DefaultProjectConfig.Labels = map[string]string{}
	modifyProjectFlags(&(project.DefaultProjectConfig.ArchiveProject), true, &(project.DefaultProjectConfig.ActivateProject), false)
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
	tearDownAndVerify(t, s.Writer, "Project dummyProject failed to update"+
		" due to Error Updating Project\n")
}

func TestEmptyProjectInput(t *testing.T) {
	s := setup()
	updateProjectSetup()
	config.GetConfig().Project = ""
	modifyProjectFlags(&(project.DefaultProjectConfig.ArchiveProject), false, &(project.DefaultProjectConfig.ActivateProject), true)
	err := updateProjectsFunc(s.Ctx, []string{}, s.CmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf(clierrors.ErrProjectNotPassed), err)
}

func TestInvalidInput(t *testing.T) {
	s := setup()
	updateProjectSetup()
	config.GetConfig().Project = projectValue
	project.DefaultProjectConfig.Name = projectValue
	modifyProjectFlags(&(project.DefaultProjectConfig.ArchiveProject), true, &(project.DefaultProjectConfig.ActivateProject), true)
	err := updateProjectsFunc(s.Ctx, []string{}, s.CmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf(clierrors.ErrInvalidStateUpdate), err)
	tearDownAndVerify(t, s.Writer, "")
}
