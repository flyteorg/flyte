package update

import (
	"context"
	"fmt"
	"os"

	"github.com/flyteorg/flytectl/clierrors"
	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/project"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	cmdUtil "github.com/flyteorg/flytectl/pkg/commandutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

const (
	projectShort = "Update project resources"
	projectLong  = `
Update the project according to the flags passed. Allows you to archive or activate a project.
Activate project flytesnacks:
::

 flytectl update project -p flytesnacks --activate

Archive project flytesnacks:

::

 flytectl update project -p flytesnacks --archive

Incorrect usage when passing both archive and activate:

::

 flytectl update project -p flytesnacks --archive --activate

Incorrect usage when passing unknown-project:

::

 flytectl update project unknown-project --archive

project ID is required flag

::

 flytectl update project unknown-project --archive

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
	edits, err := project.DefaultProjectConfig.GetProjectSpec(config.GetConfig())
	if err != nil {
		return err
	}

	if edits.Id == "" {
		return fmt.Errorf(clierrors.ErrProjectNotPassed)
	}

	currentProject, err := cmdCtx.AdminFetcherExt().GetProjectByID(ctx, edits.Id)
	if err != nil {
		return fmt.Errorf("update project %s: could not fetch project: %w", edits.Id, err)
	}

	// We do not compare currentProject against edits directly, because edits does not
	// have a complete set of project's fields - it will only contain fields that
	// the update command allows updating. (For example, it won't have Domains field
	// initialized.)
	currentProjectWithEdits := copyProjectWithEdits(currentProject, edits)
	patch, err := DiffAsYaml(diffPathBefore, diffPathAfter, currentProject, currentProjectWithEdits)
	if err != nil {
		panic(err)
	}
	if patch == "" {
		fmt.Printf("No changes detected. Skipping the update.\n")
		return nil
	}

	fmt.Printf("The following changes are to be applied.\n%s\n", patch)

	if project.DefaultProjectConfig.DryRun {
		fmt.Printf("skipping UpdateProject request (dryRun)\n")
		return nil
	}

	if !project.DefaultProjectConfig.Force && !cmdUtil.AskForConfirmation("Continue?", os.Stdin) {
		return fmt.Errorf("update aborted by user")
	}

	_, err = cmdCtx.AdminClient().UpdateProject(ctx, edits)
	if err != nil {
		return fmt.Errorf(clierrors.ErrFailedProjectUpdate, edits.Id, err)
	}

	fmt.Printf("project %s updated\n", edits.Id)
	return nil
}

// Makes a shallow copy of target and applies certain properties from edited to it.
// The properties applied are only the ones supported by update command: state, name,
// description, labels, etc.
func copyProjectWithEdits(target *admin.Project, edited *admin.Project) *admin.Project {
	copy := *target

	copy.State = edited.State
	if edited.Name != "" {
		copy.Name = edited.Name
	}
	if edited.Description != "" {
		copy.Description = edited.Description
	}
	if len(edited.GetLabels().GetValues()) != 0 {
		copy.Labels = edited.Labels
	}

	return &copy
}
