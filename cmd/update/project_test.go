package update

import (
	"errors"
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/clierrors"
	"github.com/flyteorg/flytectl/cmd/config"
	u "github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
)

const projectValue = "dummyProject"

var (
	args                 []string
	projectUpdateRequest *admin.Project
)

func updateProjectSetup() {
	mockClient = u.MockClient
	cmdCtx = u.CmdCtx
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
	setup()
	updateProjectSetup()
	modifyProjectFlags(&(DefaultProjectConfig.ArchiveProject), false, &(DefaultProjectConfig.ActivateProject), true)
	mockClient.OnUpdateProjectMatch(ctx, projectUpdateRequest).Return(nil, nil)
	err = updateProjectsFunc(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "UpdateProject", ctx, projectUpdateRequest)
	tearDownAndVerify(t, "Project dummyProject updated to ACTIVE state\n")
}

func TestActivateProjectFuncWithError(t *testing.T) {
	setup()
	updateProjectSetup()
	modifyProjectFlags(&(DefaultProjectConfig.ArchiveProject), false, &(DefaultProjectConfig.ActivateProject), true)
	mockClient.OnUpdateProjectMatch(ctx, projectUpdateRequest).Return(nil, errors.New("Error Updating Project"))
	err = updateProjectsFunc(ctx, args, cmdCtx)
	assert.NotNil(t, err)
	mockClient.AssertCalled(t, "UpdateProject", ctx, projectUpdateRequest)
	tearDownAndVerify(t, "Project dummyProject failed to get updated to ACTIVE state due to Error Updating Project\n")
}

func TestArchiveProjectFunc(t *testing.T) {
	setup()
	updateProjectSetup()
	modifyProjectFlags(&(DefaultProjectConfig.ArchiveProject), true, &(DefaultProjectConfig.ActivateProject), false)
	projectUpdateRequest = &admin.Project{
		Id:    projectValue,
		State: admin.Project_ARCHIVED,
	}
	mockClient.OnUpdateProjectMatch(ctx, projectUpdateRequest).Return(nil, nil)
	err = updateProjectsFunc(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "UpdateProject", ctx, projectUpdateRequest)
	tearDownAndVerify(t, "Project dummyProject updated to ARCHIVED state\n")
}

func TestArchiveProjectFuncWithError(t *testing.T) {
	setup()
	updateProjectSetup()
	modifyProjectFlags(&(DefaultProjectConfig.ArchiveProject), true, &(DefaultProjectConfig.ActivateProject), false)
	projectUpdateRequest = &admin.Project{
		Id:    projectValue,
		State: admin.Project_ARCHIVED,
	}
	mockClient.OnUpdateProjectMatch(ctx, projectUpdateRequest).Return(nil, errors.New("Error Updating Project"))
	err = updateProjectsFunc(ctx, args, cmdCtx)
	assert.NotNil(t, err)
	mockClient.AssertCalled(t, "UpdateProject", ctx, projectUpdateRequest)
	tearDownAndVerify(t, "Project dummyProject failed to get updated to ARCHIVED state due to Error Updating Project\n")
}

func TestEmptyProjectInput(t *testing.T) {
	setup()
	updateProjectSetup()
	config.GetConfig().Project = ""
	modifyProjectFlags(&(DefaultProjectConfig.ArchiveProject), false, &(DefaultProjectConfig.ActivateProject), true)
	mockClient.OnUpdateProjectMatch(ctx, projectUpdateRequest).Return(nil, nil)
	err = updateProjectsFunc(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertNotCalled(t, "UpdateProject", ctx, projectUpdateRequest)
	tearDownAndVerify(t, "Project not passed")
}

func TestInvalidInput(t *testing.T) {
	setup()
	updateProjectSetup()
	modifyProjectFlags(&(DefaultProjectConfig.ArchiveProject), false, &(DefaultProjectConfig.ActivateProject), false)
	mockClient.OnUpdateProjectMatch(ctx, projectUpdateRequest).Return(nil, nil)
	err = updateProjectsFunc(ctx, args, cmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf(clierrors.ErrInvalidStateUpdate), err)
	mockClient.AssertNotCalled(t, "UpdateProject", ctx, projectUpdateRequest)
	tearDownAndVerify(t, "")
}
