.. _flytectl_get_workflow:

flytectl get workflow
---------------------

Gets workflow resources

Synopsis
~~~~~~~~



Retrieves all the workflows within project and domain.(workflow,workflows can be used interchangeably in these commands)
::

 flytectl get workflow -p flytesnacks -d development

Retrieves workflow by name within project and domain.

::

 flytectl get workflow -p flytesnacks -d development  core.basic.lp.go_greet

Retrieves latest version of workflow by name within project and domain.

::

 flytectl get workflow -p flytesnacks -d development  core.basic.lp.go_greet --latest

Retrieves particular version of workflow by name within project and domain.

::

 flytectl get workflow -p flytesnacks -d development  core.basic.lp.go_greet --version v2

Retrieves all the workflows with filters.
::
 
  bin/flytectl get workflow -p flytesnacks -d development  --filter.fieldSelector="workflow.name=k8s_spark.dataframe_passing.my_smart_schema"
 
Retrieve specific workflow with filters.
::
 
  bin/flytectl get workflow -p flytesnacks -d development k8s_spark.dataframe_passing.my_smart_schema --filter.fieldSelector="workflow.version=v1"
  
Retrieves all the workflows with limit and sorting.
::
  
  bin/flytectl get -p flytesnacks -d development workflow  --filter.sortBy=created_at --filter.limit=1 --filter.asc

Retrieves all the workflow within project and domain in yaml format.

::

 flytectl get workflow -p flytesnacks -d development -o yaml

Retrieves all the workflow within project and domain in json format.

::

 flytectl get workflow -p flytesnacks -d development -o json

Visualize the graph for a workflow within project and domain in dot format.

::

 flytectl get workflow -p flytesnacks -d development  core.flyte_basics.basic_workflow.my_wf --latest -o dot

Visualize the graph for a workflow within project and domain in a dot content render.

::

 flytectl get workflow -p flytesnacks -d development  core.flyte_basics.basic_workflow.my_wf --latest -o doturl

Usage


::

  flytectl get workflow [flags]

Options
~~~~~~~

::

      --filter.asc                    Specifies the sorting order. By default flytectl sort result in descending order
      --filter.fieldSelector string   Specifies the Field selector
      --filter.limit int32            Specifies the limit (default 100)
      --filter.sortBy string          Specifies which field to sort results  (default "created_at")
  -h, --help                          help for workflow
      --latest                         flag to indicate to fetch the latest version,  version flag will be ignored in this case
      --version string                version of the workflow to be fetched.

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

