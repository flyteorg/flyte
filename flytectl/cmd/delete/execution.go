package delete

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/execution"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
)

// Long descriptions are whitespace sensitive when generating docs using Sphinx.
const (
	execCmdShort = `Terminates/deletes execution resources.`
	execCmdLong  = `
Task executions can be aborted only if they are in non-terminal state. If they are FAILED, ABORTED, or SUCCEEDED, calling terminate on them has no effect.
Terminate a single execution with its name:

::

 flytectl delete execution c6a51x2l9e  -d development  -p flytesnacks

.. note::
    The terms execution/executions are interchangeable in these commands.

Get an execution to check its state:

::

 flytectl get execution  -d development  -p flytesnacks
  ------------ ------------------------------------------------------------------------- ---------- ----------- -------------------------------- --------------- 
 | NAME (7)   | WORKFLOW NAME                                                           | TYPE     | PHASE     | STARTED                        | ELAPSED TIME  |
  ------------ ------------------------------------------------------------------------- ---------- ----------- -------------------------------- --------------- 
 | c6a51x2l9e | recipes.core.basic.lp.go_greet                                          | WORKFLOW | ABORTED   | 2021-02-17T08:13:04.680476300Z | 15.540361300s |
  ------------ ------------------------------------------------------------------------- ---------- ----------- -------------------------------- --------------- 

Terminate multiple executions with their names:
::

 flytectl delete execution eeam9s8sny p4wv4hwgc4  -d development  -p flytesnacks

Get an execution to find the state of previously terminated executions:

::

 flytectl get execution  -d development  -p flytesnacks
  ------------ ------------------------------------------------------------------------- ---------- ----------- -------------------------------- --------------- 
 | NAME (7)   | WORKFLOW NAME                                                           | TYPE     | PHASE     | STARTED                        | ELAPSED TIME  |
  ------------ ------------------------------------------------------------------------- ---------- ----------- -------------------------------- --------------- 
 | c6a51x2l9e | recipes.core.basic.lp.go_greet                                          | WORKFLOW | ABORTED   | 2021-02-17T08:13:04.680476300Z | 15.540361300s |
  ------------ ------------------------------------------------------------------------- ---------- ----------- -------------------------------- --------------- 
 | eeam9s8sny | recipes.core.basic.lp.go_greet                                          | WORKFLOW | ABORTED   | 2021-02-17T08:14:04.803084100Z | 42.306385500s |
  ------------ ------------------------------------------------------------------------- ---------- ----------- -------------------------------- --------------- 
 | p4wv4hwgc4 | recipes.core.basic.lp.go_greet                                          | WORKFLOW | ABORTED   | 2021-02-17T08:14:27.476307400Z | 19.727504400s |
  ------------ ------------------------------------------------------------------------- ---------- ----------- -------------------------------- --------------- 

Usage
`
)

func terminateExecutionFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	for i := 0; i < len(args); i++ {
		name := args[i]
		logger.Infof(ctx, "Terminating execution of %v execution ", name)
		if execution.DefaultExecDeleteConfig.DryRun {
			logger.Infof(ctx, "skipping TerminateExecution request (dryRun)")
		} else {
			_, err := cmdCtx.AdminClient().TerminateExecution(ctx, &admin.ExecutionTerminateRequest{
				Id: &core.WorkflowExecutionIdentifier{
					Project: config.GetConfig().Project,
					Domain:  config.GetConfig().Domain,
					Name:    name,
				},
			})
			if err != nil {
				logger.Errorf(ctx, "Failed to terminate execution of %v execution due to %v ", name, err)
				return err
			}
		}
		logger.Infof(ctx, "Terminated execution of %v execution ", name)
	}
	return nil
}
