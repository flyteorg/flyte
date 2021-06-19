.. _flytectl_get_task:

flytectl get task
-----------------

Gets task resources

Synopsis
~~~~~~~~



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


::

  flytectl get task [flags]

Options
~~~~~~~

::

      --execFile string                execution file name to be used for generating execution spec of a single task.
      --filter.asc                     Specifies the sorting order. By default flytectl sort result in descending order
      --filter.field-selector string   Specifies the Field selector
      --filter.limit int32             Specifies the limit (default 100)
      --filter.sort-by string          Specifies which field to sort result by 
  -h, --help                           help for task
      --latest                         flag to indicate to fetch the latest version, version flag will be ignored in this case
      --version string                 version of the task to be fetched.

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  -c, --config string    config file (default is $HOME/.flyte/config.yaml)
  -d, --domain string    Specifies the Flyte project's domain.
  -o, --output string    Specifies the output type - supported formats [TABLE JSON YAML DOT DOTURL]. NOTE: dot, doturl are only supported for Workflow (default "TABLE")
  -p, --project string   Specifies the Flyte project.

SEE ALSO
~~~~~~~~

* :doc:`flytectl_get` 	 - Used for fetching various flyte resources including tasks/workflows/launchplans/executions/project.

