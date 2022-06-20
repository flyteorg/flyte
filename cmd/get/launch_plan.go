package get

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/launchplan"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/ext"
	"github.com/flyteorg/flytectl/pkg/printer"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/golang/protobuf/proto"
)

const (
	launchPlanShort = "Gets the launch plan resources."
	launchPlanLong  = `
Retrieve all launch plans within the project and domain:
::

 flytectl get launchplan -p flytesnacks -d development

.. note::
     The terms launchplan/launchplans are interchangeable in these commands.

 Retrieve a launch plan by name within the project and domain:

::

 flytectl get launchplan -p flytesnacks -d development core.basic.lp.go_greet


Retrieve the latest version of the task by name within the project and domain:

::

 flytectl get launchplan -p flytesnacks -d development  core.basic.lp.go_greet --latest

Retrieve a particular version of the launch plan by name within the project and domain:

::

 flytectl get launchplan -p flytesnacks -d development  core.basic.lp.go_greet --version v2

Retrieve all launch plans for a given workflow name:

::

 flytectl get launchplan -p flytesnacks -d development --workflow core.flyte_basics.lp.go_greet

Retrieve all the launch plans with filters:
::

  flytectl get launchplan -p flytesnacks -d development --filter.fieldSelector="name=core.basic.lp.go_greet"

Retrieve all active launch plans:
::

  flytectl get launchplan -p flytesnacks -d development -o yaml  --filter.fieldSelector "state=1"

Retrieve all archived launch plans:
::

  flytectl get launchplan -p flytesnacks -d development -o yaml  --filter.fieldSelector "state=0"

Retrieve launch plans entity search across all versions with filters:
::

  flytectl get launchplan -p flytesnacks -d development k8s_spark.dataframe_passing.my_smart_schema --filter.fieldSelector="version=v1"


Retrieve all the launch plans with limit and sorting:
::

  flytectl get launchplan -p flytesnacks -d development --filter.sortBy=created_at --filter.limit=1 --filter.asc

Retrieve launch plans present in other pages by specifying the limit and page number:
::

  flytectl get -p flytesnacks -d development launchplan --filter.limit=10 --filter.page=2

Retrieve all launch plans within the project and domain in YAML format:

::

 flytectl get launchplan -p flytesnacks -d development -o yaml

Retrieve all launch plans the within the project and domain in JSON format:

::

 flytectl get launchplan -p flytesnacks -d development -o json

Retrieve a launch plan within the project and domain as per a version and generates the execution spec file; the file can be used to launch the execution using the 'create execution' command:

::

 flytectl get launchplan -d development -p flytectldemo core.advanced.run_merge_sort.merge_sort --execFile execution_spec.yaml

The generated file would look similar to this:

.. code-block:: yaml

	 iamRoleARN: ""
	 inputs:
	   numbers:
	   - 0
	   numbers_count: 0
	   run_local_at_count: 10
	 kubeServiceAcct: ""
	 targetDomain: ""
	 targetProject: ""
	 version: v3
	 workflow: core.advanced.run_merge_sort.merge_sort

Check the :ref:` + "`create execution section<flytectl_create_execution>`" + ` on how to launch one using the generated file.
Usage
`
)

// Column structure for get specific launchplan
var launchplanColumns = []printer.Column{
	{Header: "Version", JSONPath: "$.id.version"},
	{Header: "Name", JSONPath: "$.id.name"},
	{Header: "Type", JSONPath: "$.closure.compiledTask.template.type"},
	{Header: "State", JSONPath: "$.spec.state"},
	{Header: "Schedule", JSONPath: "$.spec.entityMetadata.schedule"},
	{Header: "Inputs", JSONPath: "$.closure.expectedInputs.parameters." + printer.DefaultFormattedDescriptionsKey + ".var.description"},
	{Header: "Outputs", JSONPath: "$.closure.expectedOutputs.variables." + printer.DefaultFormattedDescriptionsKey + ".description"},
}

// Column structure for get all launchplans
var launchplansColumns = []printer.Column{
	{Header: "Version", JSONPath: "$.id.version"},
	{Header: "Name", JSONPath: "$.id.name"},
	{Header: "Type", JSONPath: "$.id.resourceType"},
	{Header: "CreatedAt", JSONPath: "$.closure.createdAt"},
}

func LaunchplanToProtoMessages(l []*admin.LaunchPlan) []proto.Message {
	messages := make([]proto.Message, 0, len(l))
	for _, m := range l {
		messages = append(messages, m)
	}
	return messages
}

func LaunchplanToTableProtoMessages(l []*admin.LaunchPlan) []proto.Message {
	messages := make([]proto.Message, 0, len(l))
	for _, m := range l {
		m := proto.Clone(m).(*admin.LaunchPlan)
		if m.Closure != nil {
			if m.Closure.ExpectedInputs != nil && m.Closure.ExpectedInputs.Parameters != nil {
				printer.FormatParameterDescriptions(m.Closure.ExpectedInputs.Parameters)
			}
			if m.Closure.ExpectedOutputs != nil && m.Closure.ExpectedOutputs.Variables != nil {
				printer.FormatVariableDescriptions(m.Closure.ExpectedOutputs.Variables)
			}
		}
		messages = append(messages, m)
	}
	return messages
}

func getLaunchPlanFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	launchPlanPrinter := printer.Printer{}
	var launchPlans []*admin.LaunchPlan
	project := config.GetConfig().Project
	domain := config.GetConfig().Domain
	if len(args) == 1 {
		name := args[0]
		var err error
		if launchPlans, err = FetchLPForName(ctx, cmdCtx.AdminFetcherExt(), name, project, domain); err != nil {
			return err
		}
		logger.Debugf(ctx, "Retrieved %v launch plans", len(launchPlans))
		if config.GetConfig().MustOutputFormat() == printer.OutputFormatTABLE {
			err = launchPlanPrinter.Print(config.GetConfig().MustOutputFormat(), launchplanColumns,
				LaunchplanToTableProtoMessages(launchPlans)...)
		} else {
			err = launchPlanPrinter.Print(config.GetConfig().MustOutputFormat(), launchplanColumns,
				LaunchplanToProtoMessages(launchPlans)...)
		}
		if err != nil {
			return err
		}
		return nil
	}

	if len(launchplan.DefaultConfig.Workflow) > 0 {
		if len(launchplan.DefaultConfig.Filter.FieldSelector) > 0 {
			return fmt.Errorf("fieldSelector cannot be specified with workflow flag")
		}
		launchplan.DefaultConfig.Filter.FieldSelector = fmt.Sprintf("workflow.name=%s", launchplan.DefaultConfig.Workflow)
	}

	launchPlans, err := cmdCtx.AdminFetcherExt().FetchAllVerOfLP(ctx, "", config.GetConfig().Project, config.GetConfig().Domain, launchplan.DefaultConfig.Filter)
	if err != nil {
		return err
	}

	logger.Debugf(ctx, "Retrieved %v launch plans", len(launchPlans))
	if config.GetConfig().MustOutputFormat() == printer.OutputFormatTABLE {
		return launchPlanPrinter.Print(config.GetConfig().MustOutputFormat(), launchplansColumns,
			LaunchplanToTableProtoMessages(launchPlans)...)
	}
	return launchPlanPrinter.Print(config.GetConfig().MustOutputFormat(), launchplansColumns,
		LaunchplanToProtoMessages(launchPlans)...)

}

// FetchLPForName fetches the launchplan give it name.
func FetchLPForName(ctx context.Context, fetcher ext.AdminFetcherExtInterface, name, project,
	domain string) ([]*admin.LaunchPlan, error) {
	var launchPlans []*admin.LaunchPlan
	var lp *admin.LaunchPlan
	var err error
	if launchplan.DefaultConfig.Latest {
		if lp, err = fetcher.FetchLPLatestVersion(ctx, name, project, domain, launchplan.DefaultConfig.Filter); err != nil {
			return nil, err
		}
		launchPlans = append(launchPlans, lp)
	} else if launchplan.DefaultConfig.Version != "" {
		if lp, err = fetcher.FetchLPVersion(ctx, name, launchplan.DefaultConfig.Version, project, domain); err != nil {
			return nil, err
		}
		launchPlans = append(launchPlans, lp)
	} else {
		launchPlans, err = fetcher.FetchAllVerOfLP(ctx, name, project, domain, launchplan.DefaultConfig.Filter)
		if err != nil {
			return nil, err
		}
	}
	if launchplan.DefaultConfig.ExecFile != "" {
		// There would be atleast one launchplan object when code reaches here and hence the length
		// assertion is not required.
		lp = launchPlans[0]
		// Only write the first task from the tasks object.
		if err = CreateAndWriteExecConfigForWorkflow(lp, launchplan.DefaultConfig.ExecFile); err != nil {
			return nil, err
		}
	}
	return launchPlans, nil
}
