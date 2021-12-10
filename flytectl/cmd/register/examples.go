package register

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/google/go-github/github"

	rconfig "github.com/flyteorg/flytectl/cmd/config/subcommand/register"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
)

const (
	registerExampleShort = "Register Flytesnacks example"
	registerExampleLong  = `
Register all latest Flytesnacks examples:
::

 flytectl register examples  -d development  -p flytesnacks 

Register specific release of Flytesnacks examples:
::

 flytectl register examples  -d development  -p flytesnacks --version v0.2.176
	
Note: The register command automatically override the version with release version	
Usage
`
)

var (
	githubOrg             = "flyteorg"
	flytesnacksRepository = "flytesnacks"
)

func registerExamplesFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	var examples []github.ReleaseAsset

	// Deprecated checks for --k8Service
	deprecatedCheck(ctx, &rconfig.DefaultFilesConfig.K8sServiceAccount, rconfig.DefaultFilesConfig.K8ServiceAccount)

	examples, tag, err := getAllFlytesnacksExample(githubOrg, flytesnacksRepository, rconfig.DefaultFilesConfig.Version)
	if err != nil {
		return err
	}

	logger.Infof(ctx, "Register started for %s %s release https://github.com/%s/%s/releases/tag/%s", flytesnacksRepository, tag, githubOrg, flytesnacksRepository, tag)
	rconfig.DefaultFilesConfig.Archive = true
	rconfig.DefaultFilesConfig.Version = tag
	for _, v := range examples {
		args := []string{
			*v.BrowserDownloadURL,
		}
		if err := Register(ctx, args, cmdCtx); err != nil {
			return fmt.Errorf("Example %v failed to register %v", v.Name, err)
		}
	}
	return nil
}
