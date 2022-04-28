package get

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/cmd/config"
	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/workflowexecutionconfig"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	workflowExecutionConfigShort = "Gets matchable resources of workflow execution config."
	workflowExecutionConfigLong  = `
Retrieve workflow execution config for the given project and domain, in combination with the workflow name.

For project flytectldemo and development domain:

::

 flytectl get workflow-execution-config -p flytectldemo -d development 

Example: output from the command:

.. code-block:: json

 {
	"project": "flytectldemo",
	"domain": "development",
	"max_parallelism": 5
 }

Retrieve workflow execution config for the project, domain, and workflow.
For project flytectldemo, development domain and workflow 'core.control_flow.run_merge_sort.merge_sort':

::

 flytectl get workflow-execution-config -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

Example: output from the command:

.. code-block:: json

 {
	"project": "flytectldemo",
	"domain": "development",
    "workflow": "core.control_flow.run_merge_sort.merge_sort"
	"max_parallelism": 5
 }

Write the workflow execution config to a file. If there are no workflow execution config, the command throws an error.
The config file is written to wec.yaml file.
Example: content of wec.yaml:

::

 flytectl get workflow-execution-config -p flytectldemo -d development --attrFile wec.yaml


.. code-block:: yaml

    domain: development
    project: flytectldemo
    max_parallelism: 5

Generate a sample workflow execution config file to be used for creating a new workflow execution config at project domain

::
	flytectl get workflow-execution-config -p flytesnacks -d development --attrFile wec.yaml --gen


.. code-block:: yaml

	annotations:
	  values:
		cliAnnotationKey: cliAnnotationValue
	domain: development
	labels:
	  values:
		cliLabelKey: cliLabelValue
	max_parallelism: 10
	project: flytesnacks
	raw_output_data_config:
	  output_location_prefix: cliOutputLocationPrefix
	security_context:
	  run_as:
		k8s_service_account: default



Generate a sample workflow execution config file to be used for creating a new workflow execution config at project domain workflow level

::
	flytectl get workflow-execution-config -p flytesnacks -d development --attrFile wec.yaml flytectl get workflow-execution-config --gen


.. code-block:: yaml

	annotations:
	  values:
		cliAnnotationKey: cliAnnotationValue
	domain: development
	labels:
	  values:
		cliLabelKey: cliLabelValue
	max_parallelism: 10
	project: flytesnacks
	workflow: k8s_spark.dataframe_passing.my_smart_structured_dataset
	raw_output_data_config:
	  output_location_prefix: cliOutputLocationPrefix
	security_context:
	  run_as:
		k8s_service_account: default


Usage
`
)

func getWorkflowExecutionConfigFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	var project string
	var domain string
	var workflowName string

	// Get the project domain workflow name parameters from the command line. Project and domain are mandatory for this command
	project = config.GetConfig().Project
	domain = config.GetConfig().Domain
	if len(args) == 1 {
		workflowName = args[0]
	}
	// Construct a shadow config for WorkflowExecutionConfig. The shadow config is not using ProjectDomainAttribute/Workflowattribute directly inorder to simplify the inputs.
	workflowExecutionConfigFileConfig := workflowexecutionconfig.FileConfig{Project: project, Domain: domain, Workflow: workflowName}
	// Get the workflow execution config from the command line config
	fileName := workflowexecutionconfig.DefaultFetchConfig.AttrFile

	// Updates the workflowExecutionConfigFileConfig with the fetched matchable attribute
	if err := FetchAndUnDecorateMatchableAttr(ctx, project, domain, workflowName, cmdCtx.AdminFetcherExt(),
		&workflowExecutionConfigFileConfig, admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG); err != nil {
		if grpcError := status.Code(err); grpcError == codes.NotFound && workflowexecutionconfig.DefaultFetchConfig.Gen {
			fmt.Println("Generating a sample workflow execution config file")
			workflowExecutionConfigFileConfig = getSampleWorkflowExecutionFileConfig(project, domain, workflowName)
		} else {
			return err
		}
	}

	// Write the config to the file which can be used for update
	if err := sconfig.DumpTaskResourceAttr(workflowExecutionConfigFileConfig, fileName); err != nil {
		return err
	}
	return nil
}

func getSampleWorkflowExecutionFileConfig(project, domain, workflow string) workflowexecutionconfig.FileConfig {
	return workflowexecutionconfig.FileConfig{
		Project:  project,
		Domain:   domain,
		Workflow: workflow,
		WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
			MaxParallelism: 10,
			SecurityContext: &core.SecurityContext{
				RunAs: &core.Identity{
					K8SServiceAccount: "default",
					IamRole:           "",
				},
			},
			Labels: &admin.Labels{
				Values: map[string]string{"cliLabelKey": "cliLabelValue"},
			},
			Annotations: &admin.Annotations{
				Values: map[string]string{"cliAnnotationKey": "cliAnnotationValue"},
			},
			RawOutputDataConfig: &admin.RawOutputDataConfig{
				OutputLocationPrefix: "cliOutputLocationPrefix",
			},
		},
	}
}
