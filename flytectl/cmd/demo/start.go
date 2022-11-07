package demo

import (
	"context"

	"github.com/flyteorg/flytectl/pkg/sandbox"

	sandboxCmdConfig "github.com/flyteorg/flytectl/cmd/config/subcommand/sandbox"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
)

const (
	startShort = "Starts the Flyte demo cluster."
	startLong  = `
Flyte demo is a fully standalone minimal environment for running Flyte.
It provides a simplified way of running Flyte demo as a single Docker container locally.

Starts the demo cluster without any source code:
::

 flytectl demo start

Runs a dev cluster, which only has minio and postgres pod.
::

 flytectl demo start --dev

Mounts your source code repository inside the demo cluster:
::

 flytectl demo start --source=$HOME/flyteorg/flytesnacks

Specify a Flyte demo compliant image with the registry. This is useful in case you want to use an image from your registry.
::

  flytectl demo start --image docker.io/my-override:latest

Note: If image flag is passed then Flytectl will ignore version and pre flags.

Specify a Flyte demo image pull policy. Possible pull policy values are Always, IfNotPresent, or Never:
::

 flytectl demo start --image docker.io/my-override:latest --imagePullPolicy Always

Runs a specific version of Flyte. Flytectl demo only supports Flyte version available in the Github release, https://github.com/flyteorg/flyte/tags.
::

 flytectl demo start --version=v0.14.0

.. note::
	  Flytectl demo is only supported for Flyte versions >= v1.0.0

Runs the latest pre release of Flyte.
::

 flytectl demo start --pre

Start demo cluster passing environment variables. This can be used to pass docker specific env variables or flyte specific env variables.
eg : for passing timeout value in secs for the demo container use the following.
::

 flytectl demo start --env FLYTE_TIMEOUT=700

The DURATION can be a positive integer or a floating-point number, followed by an optional unit suffix::
s - seconds (default)
m - minutes
h - hours
d - days
When no unit is used, it defaults to seconds. If the duration is set to zero, the associated timeout is disabled.


eg : for passing multiple environment variables
::

 flytectl demo start --env USER=foo --env PASSWORD=bar


For just printing the docker commands for bringingup the demo container
::

 flytectl demo start --dryRun


Usage
`
)

func startDemoCluster(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	sandboxDefaultConfig := sandboxCmdConfig.DefaultConfig
	return sandbox.StartDemoCluster(ctx, args, sandboxDefaultConfig)

}
