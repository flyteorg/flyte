package sandbox

import (
	"context"

	sandboxCmdConfig "github.com/flyteorg/flytectl/cmd/config/subcommand/sandbox"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/sandbox"
)

const (
	startShort = "Starts the Flyte sandbox cluster."
	startLong  = `
Flyte sandbox is a fully standalone minimal environment for running Flyte.
It provides a simplified way of running Flyte sandbox as a single Docker container locally.

Starts the sandbox cluster without any source code:
::

 flytectl sandbox start

Mounts your source code repository inside the sandbox:

::

 flytectl sandbox start --source=$HOME/flyteorg/flytesnacks

Runs a specific version of Flyte. Flytectl sandbox only supports Flyte version available in the Github release, https://github.com/flyteorg/flyte/tags.

::

 flytectl sandbox start  --version=v0.14.0

.. note::
	  Flytectl Sandbox is only supported for Flyte versions > v0.10.0.

Runs the latest pre release of  Flyte.
::

 flytectl sandbox start  --pre

Note: The pre release flag will be ignored if the user passes the version flag. In that case, Flytectl will use a specific version.

Specify a Flyte Sandbox compliant image with the registry. This is useful in case you want to use an image from your registry.
::

  flytectl sandbox start --image docker.io/my-override:latest

Note: If image flag is passed then Flytectl will ignore version and pre flags.

Specify a Flyte Sandbox image pull policy. Possible pull policy values are Always, IfNotPresent, or Never:
::

 flytectl sandbox start  --image docker.io/my-override:latest --imagePullPolicy Always

Start sandbox cluster passing environment variables. This can be used to pass docker specific env variables or flyte specific env variables.
eg : for passing timeout value in secs for the sandbox container use the following.
::

 flytectl sandbox start --env FLYTE_TIMEOUT=700


The DURATION can be a positive integer or a floating-point number, followed by an optional unit suffix::
s - seconds (default)
m - minutes
h - hours
d - days
When no unit is used, it defaults to seconds. If the duration is set to zero, the associated timeout is disabled.


eg : for passing multiple environment variables
::

 flytectl sandbox start --env USER=foo --env PASSWORD=bar


Usage
`
)

func startSandboxCluster(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	sandboxDefaultConfig := sandboxCmdConfig.DefaultConfig
	return sandbox.StartSandboxCluster(ctx, args, sandboxDefaultConfig)
}
