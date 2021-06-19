.. _flytectl_get_plugin-override:

flytectl get plugin-override
----------------------------

Gets matchable resources of plugin override

Synopsis
~~~~~~~~



Retrieves plugin overrides for given project and domain combination or additionally with workflow name.

Retrieves plugin overrides for project and domain
Here the command get plugin override for project flytectldemo and development domain.

::

 flytectl get plugin-override -p flytectldemo -d development 

eg : output from the command

.. code-block:: json

 {
	"project": "flytectldemo",
	"domain": "development",
	"overrides": [{
		"task_type": "python_task",
		"plugin_id": ["pluginoverride1", "pluginoverride2"],
        "missing_plugin_behavior": 0 
	}]
 }

Retrieves plugin override for project and domain and workflow
Here the command get plugin override for project flytectldemo ,development domain and workflow core.control_flow.run_merge_sort.merge_sort

::

 flytectl get plugin-override -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

eg : output from the command

.. code-block:: json

 {
	"project": "flytectldemo",
	"domain": "development",
    "workflow": "core.control_flow.run_merge_sort.merge_sort"
	"overrides": [{
		"task_type": "python_task",
		"plugin_id": ["pluginoverride1", "pluginoverride2"],
        "missing_plugin_behavior": 0
	}]
 }

Writing the plugin override to a file. If there are no plugin overrides, command would return an error.
Here the command gets plugin overrides and writes the config file to po.yaml
eg:  content of po.yaml

::

 flytectl get plugin-override --attrFile po.yaml


.. code-block:: yaml

    domain: development
    project: flytectldemo
    overrides:
       - task_type: python_task # Task type for which to apply plugin implementation overrides
         plugin_id:             # Plugin id(s) to be used in place of the default for the task type.
           - plugin_override1
           - plugin_override2
         missing_plugin_behavior: 1 # Behavior when no specified plugin_id has an associated handler. 0 : FAIL , 1: DEFAULT

Usage


::

  flytectl get plugin-override [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for generating attribute for the resource type.
  -h, --help              help for plugin-override

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

