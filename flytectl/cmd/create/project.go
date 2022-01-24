package create

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/clierrors"
	"github.com/flyteorg/flytectl/cmd/config"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/project"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytestdlib/logger"
)

const (
	projectShort = "Create project resources"
	projectLong  = `
Create projects.(project/projects can be used interchangeably in these commands)

::

 flytectl create project --name flytesnacks --id flytesnacks --description "flytesnacks description"  --labels app=flyte

Create a project by definition file. Note: The name shouldn't contain any whitespace characters.
::

 flytectl create project --file project.yaml 

.. code-block:: yaml

    id: "project-unique-id"
    name: "Name"
    labels:
       values:
         app: flyte
    description: "Some description for the project"

`
)

func createProjectsCommand(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
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
	fmt.Println("project Created successfully")
	return nil
}
