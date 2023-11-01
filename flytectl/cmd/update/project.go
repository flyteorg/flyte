package update

import (
	"context"
	"fmt"
	"os"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/flyteorg/flytectl/clierrors"
	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/project"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	cmdUtil "github.com/flyteorg/flytectl/pkg/commandutils"
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
	projectConfig := project.DefaultProjectConfig

	edits, err := projectConfig.GetProjectSpec(config.GetConfig())
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
	currentProjectWithEdits := copyProjectWithEdits(currentProject, edits, projectConfig)
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
func copyProjectWithEdits(target *admin.Project, edited *admin.Project, projectConfig *project.ConfigProject) *admin.Project {
	copy := *target

	if edited.Name != "" {
		copy.Name = edited.Name
	}
	if edited.Description != "" {
		copy.Description = edited.Description
	}
	if len(edited.GetLabels().GetValues()) != 0 {
		copy.Labels = edited.Labels
	}

	// `edited` comes with `admin.Project_ACTIVE` state by default
	// if both `activate` and `archive` flags have not been set.
	//
	// This will overwrite state of `copy` if we directly set it
	// without checking for flags, which will show up on the diff.
	//
	// Also, after showing the diff, the `edited` is used for updating
	// the project, which comes with `Project_ACTIVE` by default
	// unless overwritten. Therefore, on the `else` block,
	// we overwrite the `edited` with the state of `copy`
	// if both `archive` and `activate` flags are unset.
	//
	// This is a bit hacky IMO. Proper solution would be to
	// refactor `project.ConfigProject` and this file in order to
	// separate the logic of setting `ConfigProject` struct fields
	// from creation of a 'default' project based on those flags.
	// Having a proper order of precedence between global config,
	// YAML file input, and the flags for `ConfigProject` would also
	// be good.
	if projectConfig.Archive || projectConfig.Activate {
		copy.State = edited.State
	} else {
		edited.State = copy.State
	}
	return &copy
}
