package create

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/clierrors"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/project"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytestdlib/logger"
)

const (
	projectShort = "Creates project resources."
	projectLong  = `
Create a project given its name and id.

::

 flytectl create project --name flytesnacks --id flytesnacks --description "flytesnacks description" --labels app=flyte

.. note::
   The terms project/projects are interchangeable in these commands.

Create a project by definition file.

::

 flytectl create project --file project.yaml

.. code-block:: yaml

    id: "project-unique-id"
    name: "Name"
    labels:
       values:
         app: flyte
    description: "Some description for the project."

.. note::
    The project name shouldn't contain any whitespace characters.
`
)

func createProjectsCommand(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	projectSpec, err := project.DefaultProjectConfig.GetProjectSpec(project.DefaultProjectConfig.ID)
	if err != nil {
		return err
	}
	if projectSpec.Id == "" {
		return fmt.Errorf(clierrors.ErrProjectNotPassed)
	}
	if projectSpec.Name == "" {
		return fmt.Errorf(clierrors.ErrProjectNameNotPassed)
	}

	if project.DefaultProjectConfig.DryRun {
		logger.Debugf(ctx, "skipping RegisterProject request (DryRun)")
	} else {
		_, err := cmdCtx.AdminClient().RegisterProject(ctx, &admin.ProjectRegisterRequest{
			Project: &admin.Project{
				Id:          projectSpec.Id,
				Name:        projectSpec.Name,
				Description: projectSpec.Description,
				Labels:      projectSpec.Labels,
			},
		})
		if err != nil {
			return err
		}
	}
	fmt.Println("project created successfully.")
	return nil
}
