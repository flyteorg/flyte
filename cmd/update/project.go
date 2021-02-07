package update

import (
	"context"
	"fmt"
	"github.com/lyft/flytectl/cmd/config"
	cmdCore "github.com/lyft/flytectl/cmd/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate pflags ProjectConfig

// Config hold configuration for project update flags.
type ProjectConfig struct {
	ActivateProject bool `json:"activateProject" pflag:",Activates the project specified as argument."`
	ArchiveProject  bool `json:"archiveProject" pflag:",Archives the project specified as argument."`
}

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
		fmt.Printf(errInvalidUpdate)
		return nil
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
