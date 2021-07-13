.. _flytectl_delete_workflow-execution-config:

flytectl delete workflow-execution-config
-----------------------------------------

Deletes matchable resources of workflow execution config

Synopsis
~~~~~~~~



Deletes workflow execution config for given project and domain combination or additionally with workflow name.

Deletes workflow execution config label for project and domain
Here the command delete workflow execution config for project flytectldemo and development domain.
::

 flytectl delete workflow-execution-config -p flytectldemo -d development 


Deletes workflow execution config using config file which was used for creating it.
Here the command deletes workflow execution config from the config file wec.yaml
Max_parallelism is optional in the file as its unread during the delete command but can be kept as the same file can be used for get, update or delete 
eg:  content of wec.yaml which will use the project domain and workflow name for deleting the resource

::

 flytectl delete workflow-execution-config --attrFile wec.yaml


.. code-block:: yaml
	
    domain: development
    project: flytectldemo
    max_parallelism: 5

Deletes workflow execution config for a workflow
Here the command deletes workflow execution config for a workflow core.control_flow.run_merge_sort.merge_sort

::

 flytectl delete workflow-execution-config -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

Usage


::

  flytectl delete workflow-execution-config [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for delete attribute for the resource type.
  -h, --help              help for workflow-execution-config

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  -c, --config string                              config file (default is $HOME/.flyte/config.yaml)
  -d, --domain string                              Specifies the Flyte project's domain.
  -o, --output string                              Specifies the output type - supported formats [TABLE JSON YAML DOT DOTURL]. NOTE: dot, doturl are only supported for Workflow (default "TABLE")
  -p, --project string                             Specifies the Flyte project.

SEE ALSO
~~~~~~~~

* :doc:`flytectl_delete` 	 - Used for terminating/deleting various flyte resources including tasks/workflows/launchplans/executions/project.

