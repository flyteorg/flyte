package register

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyte/flytectl/cmd/config"
	rconfig "github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/register"
	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	g "github.com/flyteorg/flyte/flytectl/pkg/github"
	"github.com/flyteorg/flyte/flytestdlib/logger"

	"github.com/google/go-github/v42/github"
)

const (
	registerExampleShort = "Registers Flytesnacks example."
	registerExampleLong  = `
Register all the latest Flytesnacks examples:
::

 flytectl register examples  -d development  -p flytesnacks 

Register specific release of Flytesnacks examples:
::

 flytectl register examples  -d development  -p flytesnacks --version v0.2.176
	
.. note::
	 The register command automatically override the version with release version.	

Usage
`
)

var (
	flytesnacks = "flytesnacks"
)

func registerExamplesFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	var examples []*github.ReleaseAsset

	// Deprecated checks for --k8Service
	deprecatedCheck(ctx, &rconfig.DefaultFilesConfig.K8sServiceAccount, rconfig.DefaultFilesConfig.K8ServiceAccount)

	ghRepo := g.GetGHRepoService()
	examples, tag, err := getAllExample(flytesnacks, rconfig.DefaultFilesConfig.Version, ghRepo)
	if err != nil {
		return err
	}

	logger.Infof(ctx, "Register started for %s %s release https://github.com/flyteorg/%s/releases/tag/%s", flytesnacks, tag, flytesnacks, tag)
	rconfig.DefaultFilesConfig.Archive = true
	rconfig.DefaultFilesConfig.Version = *tag.TagName
	for _, v := range examples {
		args := []string{
			*v.BrowserDownloadURL,
		}
		if err := Register(ctx, args, config.GetConfig(), cmdCtx); err != nil {
			return fmt.Errorf("example %v failed to register %v", v.Name, err)
		}
	}
	return nil
}
