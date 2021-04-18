package update

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/cmd/config"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate pflags ProjectConfig

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

Activates project named flytesnacks using short option -t.
::

 bin/flytectl update project -p flytesnacks -t

Archives project named flytesnacks using short option -a.

::

 bin/flytectl update project flytesnacks -a

Incorrect usage when passing both archive and activate.

::

 bin/flytectl update project flytesnacks -a -t

Incorrect usage when passing unknown-project.

::

 bin/flytectl update project unknown-project -a

Incorrect usage when passing valid project using -p option.

::

 bin/flytectl update project unknown-project -a -p known-project

Usage
`
)

var (
	projectConfig      = &ProjectConfig{}
	errProjectNotFound = "Project %v not found\n"
	errInvalidUpdate   = "Invalid state passed. Specify either activate or archive\n"
	errFailedUpdate    = "Project %v failed to get updated to %v state due to %v\n"
)

func updateProjectsFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	id := config.GetConfig().Project
	if id == "" {
		fmt.Printf(errProjectNotFound, id)
		return nil
	}
	archiveProject := projectConfig.ArchiveProject
	activateProject := projectConfig.ActivateProject
	if activateProject == archiveProject {
		return fmt.Errorf(errInvalidUpdate)
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
		fmt.Printf(errFailedUpdate, id, projectState, err)
		return nil
	}
	fmt.Printf("Project %v updated to %v state\n", id, projectState)
	return nil
}
