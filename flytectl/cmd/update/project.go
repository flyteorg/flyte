package update

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/cmd/config"

	"github.com/flyteorg/flytectl/clierrors"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/project"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytestdlib/logger"
)

const (
	projectShort = "Update project resources"
	projectLong  = `
Updates the project according to the flags passed. Allows you to archive or activate a project.
Activate project flytesnacks:
::

 flytectl update project -p flytesnacks --activate

Archive project flytesnacks:

::

 flytectl update project -p flytesnacks --archive

Incorrect usage when passing both archive and activate:

::

 flytectl update project -p flytesnacks --archiveProject --activate

Incorrect usage when passing unknown-project:

::

 flytectl update project unknown-project --archive

project ID is required flag

::

 flytectl update project unknown-project --archiveProject

Update projects.(project/projects can be used interchangeably in these commands)

::

 flytectl update project -p flytesnacks --description "flytesnacks description"  --labels app=flyte

Update a project by definition file. Note: The name shouldn't contain any whitespace characters.
::

 flytectl update project --file project.yaml 

.. code-block:: yaml

    id: "project-unique-id"
    name: "Name"
    labels:
       values:
         app: flyte
    description: "Some description for the project"

Update a project state by definition file. Note: The name shouldn't contain any whitespace characters.
::

 flytectl update project --file project.yaml  --archive

.. code-block:: yaml

    id: "project-unique-id"
    name: "Name"
    labels:
       values:
         app: flyte
    description: "Some description for the project"

Usage
`
)

func updateProjectsFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	projectSpec, err := project.DefaultProjectConfig.GetProjectSpec(config.GetConfig().Project)
	if err != nil {
		return err
	}
	if projectSpec.Id == "" {
		return fmt.Errorf(clierrors.ErrProjectNotPassed)
	}
	if projectSpec.Name == "" {
		return fmt.Errorf(clierrors.ErrProjectNameNotPassed)
	}

	state, err := project.DefaultProjectConfig.MapToAdminState()
	if err != nil {
		return err
	}
	projectSpec.State = state
	if project.DefaultProjectConfig.DryRun {
		logger.Infof(ctx, "skipping UpdateProject request (dryRun)")
	} else {
		_, err := cmdCtx.AdminClient().UpdateProject(ctx, projectSpec)
		if err != nil {
			fmt.Printf(clierrors.ErrFailedProjectUpdate, projectSpec.Id, err)
			return err
		}
	}
	fmt.Printf("Project %v updated\n", projectSpec.Id)
	return nil
}
