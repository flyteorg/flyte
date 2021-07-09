.. _flytectl_get_launchplan:

flytectl get launchplan
-----------------------

Gets launch plan resources

Synopsis
~~~~~~~~



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


::

  flytectl get launchplan [flags]

Options
~~~~~~~

::

      --execFile string               execution file name to be used for generating execution spec of a single launchplan.
      --filter.asc                    Specifies the sorting order. By default flytectl sort result in descending order
      --filter.fieldSelector string   Specifies the Field selector
      --filter.limit int32            Specifies the limit (default 100)
      --filter.sortBy string          Specifies which field to sort results  (default "created_at")
  -h, --help                          help for launchplan
      --latest                         flag to indicate to fetch the latest version,  version flag will be ignored in this case
      --version string                version of the launchplan to be fetched.

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

