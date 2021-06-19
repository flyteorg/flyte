.. _flytectl_delete_plugin-override:

flytectl delete plugin-override
-------------------------------

Deletes matchable resources of plugin overrides

Synopsis
~~~~~~~~



Deletes plugin override for given project and domain combination or additionally with workflow name.

Deletes plugin override for project and domain
Here the command deletes plugin override for project flytectldemo and development domain.
::

 flytectl delete plugin-override -p flytectldemo -d development 


Deletes plugin override using config file which was used for creating it.
Here the command deletes plugin overrides from the config file po.yaml
Overrides are optional in the file as they are unread during the delete command but can be kept as the same file can be used for get, update or delete 
eg:  content of po.yaml which will use the project domain and workflow name for deleting the resource

::

 flytectl delete plugin-override --attrFile po.yaml


.. code-block:: yaml

    domain: development
    project: flytectldemo
    overrides:
       - task_type: python_task # Task type for which to apply plugin implementation overrides
         plugin_id:             # Plugin id(s) to be used in place of the default for the task type.
           - plugin_override1
           - plugin_override2
         missing_plugin_behavior: 1 # Behavior when no specified plugin_id has an associated handler. 0 : FAIL , 1: DEFAULT

Deletes plugin override for a workflow
Here the command deletes the plugin override for a workflow core.control_flow.run_merge_sort.merge_sort

::

 flytectl delete plugin-override -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

Usage


::

  flytectl delete plugin-override [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for delete attribute for the resource type.
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

* :doc:`flytectl_delete` 	 - Used for terminating/deleting various flyte resources including tasks/workflows/launchplans/executions/project.

