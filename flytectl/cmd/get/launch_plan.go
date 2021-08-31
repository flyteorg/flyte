package get

import (
	"context"

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
	launchPlanShort = "Gets launch plan resources"
	launchPlanLong  = `
Retrieves all the launch plans within project and domain.(launchplan,launchplans can be used interchangeably in these commands)
::

 flytectl get launchplan -p flytesnacks -d development

Retrieves launch plan by name within project and domain.

::

 flytectl get launchplan -p flytesnacks -d development core.basic.lp.go_greet


Retrieves latest version of task by name within project and domain.

::

 flytectl get launchplan -p flytesnacks -d development  core.basic.lp.go_greet --latest

Retrieves particular version of launchplan by name within project and domain.

::

 flytectl get launchplan -p flytesnacks -d development  core.basic.lp.go_greet --version v2

Retrieves all the launch plans with filters.
::
 
  bin/flytectl get launchplan -p flytesnacks -d development --filter.fieldSelector="name=core.basic.lp.go_greet"
 
Retrieves launch plans entity search across all versions with filters.
::
 
  bin/flytectl get launchplan -p flytesnacks -d development k8s_spark.dataframe_passing.my_smart_schema --filter.fieldSelector="version=v1"
 
 
Retrieves all the launch plans with limit and sorting.
::
 
  bin/flytectl get launchplan -p flytesnacks -d development --filter.sortBy=created_at --filter.limit=1 --filter.asc
 

Retrieves all the launchplan within project and domain in yaml format.

::

 flytectl get launchplan -p flytesnacks -d development -o yaml

Retrieves all the launchplan within project and domain in json format

::

 flytectl get launchplan -p flytesnacks -d development -o json

Retrieves a launch plans within project and domain for a version and generate the execution spec file for it to be used for launching the execution using create execution.

::

 flytectl get launchplan -d development -p flytectldemo core.advanced.run_merge_sort.merge_sort --execFile execution_spec.yaml

The generated file would look similar to this

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

Check the create execution section on how to launch one using the generated file.

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
			if m.Closure.ExpectedInputs != nil {
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
