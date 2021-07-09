.. _flytectl_get_execution:

flytectl get execution
----------------------

Gets execution resources

Synopsis
~~~~~~~~



Retrieves all the executions within project and domain.(execution,executions can be used interchangeably in these commands)
::

 bin/flytectl get execution -p flytesnacks -d development

Retrieves execution by name within project and domain.

::

 bin/flytectl get execution -p flytesnacks -d development oeh94k9r2r

Retrieves all the executions with filters.
::
 
  bin/flytectl get execution -p flytesnacks -d development --filter.fieldSelector="execution.phase in (FAILED;SUCCEEDED),execution.duration<200" 

 
Retrieves all the execution with limit and sorting.
::
  
   bin/flytectl get execution -p flytesnacks -d development --filter.sortBy=created_at --filter.limit=1 --filter.asc
   

Retrieves all the execution within project and domain in yaml format

::

 bin/flytectl get execution -p flytesnacks -d development -o yaml

Retrieves all the execution within project and domain in json format.

::

 bin/flytectl get execution -p flytesnacks -d development -o json


Get more details for the execution using --details flag which shows node executions along with task executions on them. Default view is tree view and TABLE format is not supported on this view

::

 bin/flytectl get execution -p flytesnacks -d development oeh94k9r2r --details

Using yaml view for the details. In this view only node details are available. For task details pass --nodeId flag

::

 bin/flytectl get execution -p flytesnacks -d development oeh94k9r2r --details -o yaml

Using --nodeId flag to get task executions on a specific node. Use the nodeId attribute from node details view

::

 bin/flytectl get execution -p flytesnacks -d development oeh94k9r2r --nodId n0

Task execution view is also available in yaml/json format. Below example shows yaml

::

 bin/flytectl get execution -p flytesnacks -d development oeh94k9r2r --nodId n0 -o yaml

Usage


::

  flytectl get execution [flags]

Options
~~~~~~~

::

      --details                       gets node execution details. Only applicable for single execution name i.e get execution name --details
      --filter.asc                    Specifies the sorting order. By default flytectl sort result in descending order
      --filter.fieldSelector string   Specifies the Field selector
      --filter.limit int32            Specifies the limit (default 100)
      --filter.sortBy string          Specifies which field to sort results  (default "created_at")
  -h, --help                          help for execution
      --nodeId string                 get task executions for given node name.

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

