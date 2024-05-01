package create

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/logger"

	"github.com/flyteorg/flytectl/cmd/config"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
)

const (
	executionShort = "Creates execution resources."
	executionLong  = `
Create execution resources for a given workflow or task in a project and domain.

There are three steps to generate an execution, as outlined below:

1. Generate the execution spec file using the :ref:` + "`get task <flytectl_get_task>`" + ` command.
::

	flytectl get tasks -d development -p flytesnacks core.control_flow.merge_sort.merge --version v2 --execFile execution_spec.yaml

The generated file would look similar to the following:

.. code-block:: yaml

	iamRoleARN: ""
	inputs:
	sorted_list1:
	- 0
	sorted_list2:
	- 0
	kubeServiceAcct: ""
	targetDomain: ""
	targetProject: ""
	task: core.control_flow.merge_sort.merge
	version: "v2"

2. [Optional] Update the inputs for the execution, if needed.
The generated spec file can be modified to change the input values, as shown below:

.. code-block:: yaml

	iamRoleARN: 'arn:aws:iam::12345678:role/defaultrole'
	inputs:
	sorted_list1:
	- 2
	- 4
	- 6
	sorted_list2:
	- 1
	- 3
	- 5
	kubeServiceAcct: ""
	targetDomain: ""
	targetProject: ""
	task: core.control_flow.merge_sort.merge
	version: "v2"

3. [Optional] Update the envs for the execution, if needed.
The generated spec file can be modified to change the envs values, as shown below:

.. code-block:: yaml

    iamRoleARN: ""
    inputs:
    sorted_list1:
    - 0
    sorted_list2:
    - 0
    envs:
      foo: bar
    kubeServiceAcct: ""
    targetDomain: ""
    targetProject: ""
    task: core.control_flow.merge_sort.merge
    version: "v2"

4. Run the execution by passing the generated YAML file.
The file can then be passed through the command line.
It is worth noting that the source's and target's project and domain can be different.
::

	flytectl create execution --execFile execution_spec.yaml -p flytesnacks -d staging --targetProject flytesnacks

5. To relaunch an execution, pass the current execution ID as follows:

::

 flytectl create execution --relaunch ffb31066a0f8b4d52b77 -p flytesnacks -d development

6. To recover an execution, i.e., recreate it from the last known failure point for previously-run workflow execution, run:

::

 flytectl create execution --recover ffb31066a0f8b4d52b77 -p flytesnacks -d development

See :ref:` + "`ref_flyteidl.admin.ExecutionRecoverRequest`" + ` for more details.

7. You can create executions idempotently by naming them. This is also a way to *name* an execution for discovery. Note,
an execution id has to be unique within a project domain. So if the *name* matches an existing execution an already exists exceptioj
will be raised.

::

   flytectl create execution --recover ffb31066a0f8b4d52b77 -p flytesnacks -d development custom_name

8. Generic/Struct/Dataclass/JSON types are supported for execution in a similar manner.
The following is an example of how generic data can be specified while creating the execution.

::

 flytectl get task -d development -p flytesnacks  core.type_system.custom_objects.add --execFile adddatanum.yaml

The generated file would look similar to this. Here, empty values have been dumped for generic data types 'x' and 'y'.
::

    iamRoleARN: ""
    inputs:
      "x": {}
      "y": {}
    kubeServiceAcct: ""
    targetDomain: ""
    targetProject: ""
    task: core.type_system.custom_objects.add
    version: v3

9. Modified file with struct data populated for 'x' and 'y' parameters for the task "core.type_system.custom_objects.add":

::

  iamRoleARN: "arn:aws:iam::123456789:role/dummy"
  inputs:
    "x":
      "x": 2
      "y": ydatafory
      "z":
        1 : "foo"
        2 : "bar"
    "y":
      "x": 3
      "y": ydataforx
      "z":
        3 : "buzz"
        4 : "lightyear"
  kubeServiceAcct: ""
  targetDomain: ""
  targetProject: ""
  task: core.type_system.custom_objects.add
  version: v3

10. If you have configured a plugin that implements github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces/WorkflowExecutor 
   that supports cluster pools, then when creating a new execution, you can assign it to a specific cluster pool:

::

   flytectl create execution --execFile execution_spec.yaml -p flytesnacks -d development --clusterPool my-gpu-cluster
`
)

//go:generate pflags ExecutionConfig --default-var executionConfig --bind-default-var

// ExecutionConfig hold configuration for create execution flags and configuration of the actual task or workflow  to be launched.
type ExecutionConfig struct {
	// pflag section
	ExecFile        string `json:"execFile,omitempty" pflag:",file for the execution params. If not specified defaults to <<workflow/task>_name>.execution_spec.yaml"`
	TargetDomain    string `json:"targetDomain" pflag:",project where execution needs to be created. If not specified configured domain would be used."`
	TargetProject   string `json:"targetProject" pflag:",project where execution needs to be created. If not specified configured project would be used."`
	KubeServiceAcct string `json:"kubeServiceAcct" pflag:",kubernetes service account AuthRole for launching execution."`
	IamRoleARN      string `json:"iamRoleARN" pflag:",iam role ARN AuthRole for launching execution."`
	Relaunch        string `json:"relaunch" pflag:",execution id to be relaunched."`
	Recover         string `json:"recover" pflag:",execution id to be recreated from the last known failure point."`
	DryRun          bool   `json:"dryRun" pflag:",execute command without making any modifications."`
	Version         string `json:"version" pflag:",specify version of execution workflow/task."`
	ClusterPool     string `json:"clusterPool" pflag:",specify which cluster pool to assign execution to."`
	OverwriteCache  bool   `json:"overwriteCache" pflag:",skip cached results when performing execution,causing all outputs to be re-calculated and stored data to be overwritten. Does not work for recovered executions."`
	// Non plfag section is read from the execution config generated by get task/launch plan
	Workflow string                 `json:"workflow,omitempty"`
	Task     string                 `json:"task,omitempty"`
	Inputs   map[string]interface{} `json:"inputs" pflag:"-"`
	Envs     map[string]string      `json:"envs" pflag:"-"`
}

type ExecutionType int

const (
	Task ExecutionType = iota
	Workflow
	Relaunch
	Recover
)

type ExecutionParams struct {
	name     string
	execType ExecutionType
}

var executionConfig = &ExecutionConfig{}

func createExecutionCommand(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	sourceProject := config.GetConfig().Project
	sourceDomain := config.GetConfig().Domain

	var targetExecName string
	if len(args) > 0 {
		targetExecName = args[0]
	}

	execParams, err := readConfigAndValidate(config.GetConfig().Project, config.GetConfig().Domain)
	if err != nil {
		return err
	}
	var executionRequest *admin.ExecutionCreateRequest
	switch execParams.execType {
	case Relaunch:
		return relaunchExecution(ctx, execParams.name, sourceProject, sourceDomain, cmdCtx, executionConfig, targetExecName)
	case Recover:
		return recoverExecution(ctx, execParams.name, sourceProject, sourceDomain, cmdCtx, executionConfig, targetExecName)
	case Task:
		executionRequest, err = createExecutionRequestForTask(ctx, execParams.name, sourceProject, sourceDomain, cmdCtx, executionConfig, targetExecName)
		if err != nil {
			return err
		}
	case Workflow:
		executionRequest, err = createExecutionRequestForWorkflow(ctx, execParams.name, sourceProject, sourceDomain, cmdCtx, executionConfig, targetExecName)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid execution type %v", execParams.execType)
	}

	if executionConfig.DryRun {
		logger.Debugf(ctx, "skipping CreateExecution request (DryRun)")
	} else {
		exec, _err := cmdCtx.AdminClient().CreateExecution(ctx, executionRequest)
		if _err != nil {
			return _err
		}
		fmt.Printf("execution identifier %v\n", exec.Id)
	}
	return nil
}
