.. _flytectl_get_workflow-execution-config:

flytectl get workflow-execution-config
--------------------------------------

Gets matchable resources of workflow execution config

Synopsis
~~~~~~~~



Retrieves workflow execution config for given project and domain combination or additionally with workflow name.

Retrieves workflow execution config for project and domain
Here the command get workflow execution config for project flytectldemo and development domain.

::

 flytectl get workflow-execution-config -p flytectldemo -d development 

eg : output from the command

.. code-block:: json

 {
	"project": "flytectldemo",
	"domain": "development",
	"max_parallelism": 5
 }

Retrieves workflow execution config for project and domain and workflow
Here the command get workflow execution config for project flytectldemo ,development domain and workflow core.control_flow.run_merge_sort.merge_sort

::

 flytectl get workflow-execution-config -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

eg : output from the command

.. code-block:: json

 {
	"project": "flytectldemo",
	"domain": "development",
    "workflow": "core.control_flow.run_merge_sort.merge_sort"
	"max_parallelism": 5
 }

Writing the workflow execution config to a file. If there are no workflow execution config, command would return an error.
Here the command gets workflow execution config and writes the config file to wec.yaml
eg:  content of wec.yaml

::

 flytectl get workflow-execution-config -p flytectldemo -d development --attrFile wec.yaml


.. code-block:: yaml

    domain: development
    project: flytectldemo
    max_parallelism: 5

Usage


::

  flytectl get workflow-execution-config [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for generating attribute for the resource type.
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

* :doc:`flytectl_get` 	 - Used for fetching various flyte resources including tasks/workflows/launchplans/executions/project.

