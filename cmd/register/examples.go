package register

import (
	"context"
	"fmt"

	rconfig "github.com/flyteorg/flytectl/cmd/config/subcommand/register"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
)

const (
	registerExampleShort = "Registers flytesnack example"
	registerExampleLong  = `
Registers all latest flytesnacks example
::

 bin/flytectl register examples  -d development  -p flytesnacks


Usage
`
	githubOrg        = "flyteorg"
	githubRepository = "flytesnacks"
	archive          = true
	snackReleaseURL  = "https://github.com/flyteorg/flytesnacks/releases/download/%s/flytesnacks-%s.tgz"
	flyteManifest    = "https://github.com/flyteorg/flytesnacks/releases/download/%s/flyte_tests_manifest.json"
)

func registerExamplesFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	flytesnacks, tag, err := getFlyteTestManifest()
	if err != nil {
		return err
	}
	rconfig.DefaultFilesConfig.Archive = archive
	for _, v := range flytesnacks {
		args := []string{
			fmt.Sprintf(snackReleaseURL, tag, v.Name),
		}
		if err := Register(ctx, args, cmdCtx); err != nil {
			return fmt.Errorf("Example %v failed to register %v", v.Name, err)
		}
	}
	return nil
}
