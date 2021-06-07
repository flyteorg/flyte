package get

import (
	"context"

	"github.com/flyteorg/flytectl/cmd/config"
	taskConfig "github.com/flyteorg/flytectl/cmd/config/subcommand/task"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
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

Retrieves latest version of task by name within project and domain.

::

 flytectl get task -p flytesnacks -d development  core.basic.lp.greet --latest

Retrieves particular version of task by name within project and domain.

::

 flytectl get task -p flytesnacks -d development  core.basic.lp.greet --version v2

Retrieves all the tasks with filters.
::
  
  bin/flytectl get task -p flytesnacks -d development --filter.field-selector="task.name=k8s_spark.pyspark_pi.print_every_time,task.version=v1" 
 
Retrieve a specific task with filters.
::
 
  bin/flytectl get task -p flytesnacks -d development k8s_spark.pyspark_pi.print_every_time --filter.field-selector="task.version=v1,created_at>=2021-05-24T21:43:12.325335Z" 
  
Retrieves all the task with limit and sorting.
::
   
  bin/flytectl get -p flytesnacks -d development task  --filter.sort-by=created_at --filter.limit=1 --filter.asc

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
	var tasks []*admin.Task
	var err error
	project := config.GetConfig().Project
	domain := config.GetConfig().Domain
	if len(args) == 1 {
		name := args[0]
		if tasks, err = FetchTaskForName(ctx, cmdCtx.AdminFetcherExt(), name, project, domain); err != nil {
			return err
		}
		logger.Debugf(ctx, "Retrieved Task", tasks)
		return taskPrinter.Print(config.GetConfig().MustOutputFormat(), taskColumns, TaskToProtoMessages(tasks)...)
	}
	tasks, err = cmdCtx.AdminFetcherExt().FetchAllVerOfTask(ctx, "", config.GetConfig().Project, config.GetConfig().Domain, taskConfig.DefaultConfig.Filter)
	if err != nil {
		return err
	}
	logger.Debugf(ctx, "Retrieved %v Task", len(tasks))
	return taskPrinter.Print(config.GetConfig().MustOutputFormat(), taskColumns, TaskToProtoMessages(tasks)...)
}

// FetchTaskForName Reads the task config to drive fetching the correct tasks.
func FetchTaskForName(ctx context.Context, fetcher ext.AdminFetcherExtInterface, name, project, domain string) ([]*admin.Task, error) {
	var tasks []*admin.Task
	var err error
	var task *admin.Task
	if taskConfig.DefaultConfig.Latest {
		if task, err = fetcher.FetchTaskLatestVersion(ctx, name, project, domain, taskConfig.DefaultConfig.Filter); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	} else if taskConfig.DefaultConfig.Version != "" {
		if task, err = fetcher.FetchTaskVersion(ctx, name, taskConfig.DefaultConfig.Version, project, domain); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	} else {
		tasks, err = fetcher.FetchAllVerOfTask(ctx, name, project, domain, taskConfig.DefaultConfig.Filter)
		if err != nil {
			return nil, err
		}
	}
	if taskConfig.DefaultConfig.ExecFile != "" {
		// There would be atleast one task object when code reaches here and hence the length assertion is not required.
		task = tasks[0]
		// Only write the first task from the tasks object.
		if err = CreateAndWriteExecConfigForTask(task, taskConfig.DefaultConfig.ExecFile); err != nil {
			return nil, err
		}
	}
	return tasks, nil
}
