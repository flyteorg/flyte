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
	projectShort = "Update the characteristics of a project"
	projectLong  = `
Allows you to update the characteristics of a project, including its name, labels and description.
Also allows you to archive or activate (unarchive) a project.

To archive a project, specify its ID with the *p* flag and add the *archive* flag:

::

    flytectl update project -p my-project-id --archive

To activate (unarchive) an archived project, specify its ID with the *p* flag and add the *activate* flag:

::

	flytectl update project -p my-project-id --activate

To update the characteristics of a project using flags, specify the project ID with the *p* flag and the flags corresponding to the characteristics you want to update:

::

    flytectl update project -p my-project-id --description "A wonderful project"  --labels app=my-app

To update the characteristics of a project using a *yaml* file, define the file with the project ID desired updates:

.. code-block:: yaml

    id: "my-project-id"
    name: "my-project-name"
    labels:
       values:
         app: my-app
    description: "A wonderful project"


(Note: The name parameter must not contain whitespace)

Then, pass it in using the *file* flag:

::

    flytectl update project --file project.yaml

To archive or activate (unarchive) a project using a *yaml* file:

* Add a state field, with a value of *0* for activated (unarchived) or *1* for archived, at the top level of the the *yaml* file.

* Add the *archive* flag to the command.

For example, to archive a project:

.. code-block:: yaml

    # update.yaml
    id: "my-project-id"
    state: 1

::

    $ uctl update project --file update.yaml --archive

And to activate (unarchive) the same project:

.. code-block:: yaml

    # update.yaml
    id: "my-project-id"
    state: 0

::

    $ uctl update project --file update.yaml --archive

Note that when using a *yaml* file, the *activate* flag is not used.
Instead, the *archive* flag is used for *both* archiving and activating (unarchiving) with the difference being in the *state* field of the *yaml* file.
Furthermore, the *state* field only takes effect if the *archive* flag is present in the command.

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
