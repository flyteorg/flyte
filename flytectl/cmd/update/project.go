package update

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/clierrors"

	"github.com/flyteorg/flytectl/cmd/config"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate pflags ProjectConfig --default-var DefaultProjectConfig --bind-default-var

// Config hold configuration for project update flags.
type ProjectConfig struct {
	ActivateProject bool `json:"activateProject" pflag:",Activates the project specified as argument."`
	ArchiveProject  bool `json:"archiveProject" pflag:",Archives the project specified as argument."`
}

const (
	projectShort = "Updates project resources"
	projectLong  = `
Updates the project according the flags passed. Allows you to archive or activate a project.
Activates project named flytesnacks.
::

 bin/flytectl update project -p flytesnacks --activateProject

Archives project named flytesnacks.

::

 bin/flytectl update project -p flytesnacks --archiveProject

Incorrect usage when passing both archive and activate.

::

 bin/flytectl update project flytesnacks --archiveProject --activateProject

Incorrect usage when passing unknown-project.

::

 bin/flytectl update project unknown-project --archiveProject

Incorrect usage when passing valid project using -p option.

::

 bin/flytectl update project unknown-project --archiveProject -p known-project

Usage
`
)

var DefaultProjectConfig = &ProjectConfig{}

func updateProjectsFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	id := config.GetConfig().Project
	if id == "" {
		fmt.Printf(clierrors.ErrProjectNotPassed)
		return nil
	}
	archiveProject := DefaultProjectConfig.ArchiveProject
	activateProject := DefaultProjectConfig.ActivateProject
	if activateProject == archiveProject {
		return fmt.Errorf(clierrors.ErrInvalidStateUpdate)
	}
	projectState := admin.Project_ACTIVE
	if archiveProject {
		projectState = admin.Project_ARCHIVED
	}
	_, err := cmdCtx.AdminClient().UpdateProject(ctx, &admin.Project{
		Id:    id,
		State: projectState,
	})
	if err != nil {
		fmt.Printf(clierrors.ErrFailedProjectUpdate, id, projectState, err)
		return err
	}
	fmt.Printf("Project %v updated to %v state\n", id, projectState)
	return nil
}
