package get

import (
	"context"

	"github.com/flyteorg/flytectl/cmd/config"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/adminutils"
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

Retrieves launchplan by filters.
::

 Not yet implemented

Retrieves all the launchplan within project and domain in yaml format.

::

 flytectl get launchplan -p flytesnacks -d development -o yaml

Retrieves all the launchplan within project and domain in json format

::

 flytectl get launchplan -p flytesnacks -d development -o json

Retrieves a launch plans within project and domain for a version and generate the execution spec file for it to be used for launching the execution using create execution.

::

 flytectl get launchplan -d development -p flytectldemo core.advanced.run_merge_sort.merge_sort --execFile execution_spec.yam

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

//go:generate pflags LaunchPlanConfig --default-var launchPlanConfig
var (
	launchPlanConfig = &LaunchPlanConfig{}
)

// LaunchPlanConfig
type LaunchPlanConfig struct {
	ExecFile string `json:"execFile" pflag:",execution file name to be used for generating execution spec of a single launchplan."`
	Version  string `json:"version" pflag:",version of the launchplan to be fetched."`
	Latest   bool   `json:"latest" pflag:", flag to indicate to fetch the latest version, version flag will be ignored in this case"`
}

var launchplanColumns = []printer.Column{
	{Header: "Version", JSONPath: "$.id.version"},
	{Header: "Name", JSONPath: "$.id.name"},
	{Header: "Type", JSONPath: "$.closure.compiledTask.template.type"},
	{Header: "State", JSONPath: "$.spec.state"},
	{Header: "Schedule", JSONPath: "$.spec.entityMetadata.schedule"},
}

func LaunchplanToProtoMessages(l []*admin.LaunchPlan) []proto.Message {
	messages := make([]proto.Message, 0, len(l))
	for _, m := range l {
		messages = append(messages, m)
	}
	return messages
}

func getLaunchPlanFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	launchPlanPrinter := printer.Printer{}
	project := config.GetConfig().Project
	domain := config.GetConfig().Domain
	if len(args) == 1 {
		name := args[0]
		var launchPlans []*admin.LaunchPlan
		var err error
		if launchPlans, err = FetchLPForName(ctx, cmdCtx.AdminFetcherExt(), name, project, domain); err != nil {
			return err
		}
		logger.Debugf(ctx, "Retrieved %v launch plans", len(launchPlans))
		err = launchPlanPrinter.Print(config.GetConfig().MustOutputFormat(), launchplanColumns,
			LaunchplanToProtoMessages(launchPlans)...)
		if err != nil {
			return err
		}
		return nil
	}

	launchPlans, err := adminutils.GetAllNamedEntities(ctx, cmdCtx.AdminClient().ListLaunchPlanIds,
		adminutils.ListRequest{Project: project, Domain: domain})
	if err != nil {
		return err
	}
	logger.Debugf(ctx, "Retrieved %v launch plans", len(launchPlans))
	return launchPlanPrinter.Print(config.GetConfig().MustOutputFormat(), entityColumns,
		adminutils.NamedEntityToProtoMessage(launchPlans)...)
}

// FetchLPForName fetches the launchplan give it name.
func FetchLPForName(ctx context.Context, fetcher ext.AdminFetcherExtInterface, name, project,
	domain string) ([]*admin.LaunchPlan, error) {
	var launchPlans []*admin.LaunchPlan
	var lp *admin.LaunchPlan
	var err error
	if launchPlanConfig.Latest {
		if lp, err = fetcher.FetchLPLatestVersion(ctx, name, project, domain); err != nil {
			return nil, err
		}
		launchPlans = append(launchPlans, lp)
	} else if launchPlanConfig.Version != "" {
		if lp, err = fetcher.FetchLPVersion(ctx, name, launchPlanConfig.Version, project, domain); err != nil {
			return nil, err
		}
		launchPlans = append(launchPlans, lp)
	} else {
		launchPlans, err = fetcher.FetchAllVerOfLP(ctx, name, project, domain)
		if err != nil {
			return nil, err
		}
	}
	if launchPlanConfig.ExecFile != "" {
		// There would be atleast one launchplan object when code reaches here and hence the length
		// assertion is not required.
		lp = launchPlans[0]
		// Only write the first task from the tasks object.
		if err = CreateAndWriteExecConfigForWorkflow(lp, launchPlanConfig.ExecFile); err != nil {
			return nil, err
		}
	}
	return launchPlans, nil
}
