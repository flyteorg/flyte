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
	taskShort = "Gets task resources"
	taskLong  = `
Retrieves all the task within project and domain.(task,tasks can be used interchangeably in these commands)
::

 bin/flytectl get task -p flytesnacks -d development

Retrieves task by name within project and domain.

::

 bin/flytectl task -p flytesnacks -d development core.basic.lp.greet

Retrieves project by filters.
::

 Not yet implemented

Retrieves all the tasks within project and domain in yaml format.

::

 bin/flytectl get task -p flytesnacks -d development -o yaml

Retrieves all the tasks within project and domain in json format.

::

 bin/flytectl get task -p flytesnacks -d development -o json

Retrieves a tasks within project and domain for a version and generate the execution spec file for it to be used for launching the execution using create execution.

::

 bin/flytectl get tasks -d development -p flytesnacks core.advanced.run_merge_sort.merge --execFile execution_spec.yaml --version v2

The generated file would look similar to this

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
	 task: core.advanced.run_merge_sort.merge
	 version: v2

Check the create execution section on how to launch one using the generated file.

Usage
`
)

//go:generate pflags TaskConfig --default-var taskConfig
var (
	taskConfig = &TaskConfig{}
)

// FilesConfig
type TaskConfig struct {
	ExecFile string `json:"execFile" pflag:",execution file name to be used for generating execution spec of a single task."`
	Version  string `json:"version" pflag:",version of the task to be fetched."`
	Latest   bool   `json:"latest" pflag:", flag to indicate to fetch the latest version, version flag will be ignored in this case"`
}

var taskColumns = []printer.Column{
	{Header: "Version", JSONPath: "$.id.version"},
	{Header: "Name", JSONPath: "$.id.name"},
	{Header: "Type", JSONPath: "$.closure.compiledTask.template.type"},
	{Header: "Discoverable", JSONPath: "$.closure.compiledTask.template.metadata.discoverable"},
	{Header: "Discovery Version", JSONPath: "$.closure.compiledTask.template.metadata.discoveryVersion"},
	{Header: "Created At", JSONPath: "$.closure.createdAt"},
}

func TaskToProtoMessages(l []*admin.Task) []proto.Message {
	messages := make([]proto.Message, 0, len(l))
	for _, m := range l {
		messages = append(messages, m)
	}
	return messages
}

func getTaskFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	taskPrinter := printer.Printer{}
	project := config.GetConfig().Project
	domain := config.GetConfig().Domain
	if len(args) == 1 {
		name := args[0]
		var tasks []*admin.Task
		var err error
		if tasks, err = FetchTaskForName(ctx, cmdCtx.AdminFetcherExt(), name, project, domain); err != nil {
			return err
		}
		logger.Debugf(ctx, "Retrieved Task", tasks)
		return taskPrinter.Print(config.GetConfig().MustOutputFormat(), taskColumns, TaskToProtoMessages(tasks)...)
	}
	tasks, err := adminutils.GetAllNamedEntities(ctx, cmdCtx.AdminClient().ListTaskIds, adminutils.ListRequest{Project: project, Domain: domain})
	if err != nil {
		return err
	}
	logger.Debugf(ctx, "Retrieved %v Task", len(tasks))
	return taskPrinter.Print(config.GetConfig().MustOutputFormat(), entityColumns, adminutils.NamedEntityToProtoMessage(tasks)...)
}

// FetchTaskForName Reads the task config to drive fetching the correct tasks.
func FetchTaskForName(ctx context.Context, fetcher ext.AdminFetcherExtInterface, name, project, domain string) ([]*admin.Task, error) {
	var tasks []*admin.Task
	var err error
	var task *admin.Task
	if taskConfig.Latest {
		if task, err = fetcher.FetchTaskLatestVersion(ctx, name, project, domain); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	} else if taskConfig.Version != "" {
		if task, err = fetcher.FetchTaskVersion(ctx, name, taskConfig.Version, project, domain); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	} else {
		tasks, err = fetcher.FetchAllVerOfTask(ctx, name, project, domain)
		if err != nil {
			return nil, err
		}
	}
	if taskConfig.ExecFile != "" {
		// There would be atleast one task object when code reaches here and hence the length assertion is not required.
		task = tasks[0]
		// Only write the first task from the tasks object.
		if err = CreateAndWriteExecConfigForTask(task, taskConfig.ExecFile); err != nil {
			return nil, err
		}
	}
	return tasks, nil
}
