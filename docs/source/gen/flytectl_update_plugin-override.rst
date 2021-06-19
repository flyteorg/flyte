.. _flytectl_update_plugin-override:

flytectl update plugin-override
-------------------------------

Updates matchable resources of plugin overrides

Synopsis
~~~~~~~~



Updates plugin overrides for given project and domain combination or additionally with workflow name.

Updating to the plugin override is only available from a generated file. See the get section for generating this file.
Also this will completely overwrite any existing plugins overrides on custom project and domain and workflow combination.
Would be preferable to do get and generate an plugin override file if there is an existing override already set and then update it to have new values
Refer to get plugin-override section on how to generate this file
Here the command updates takes the input for plugin overrides from the config file po.yaml
eg:  content of po.yaml

.. code-block:: yaml

    domain: development
    project: flytectldemo
    overrides:
       - task_type: python_task # Task type for which to apply plugin implementation overrides
         plugin_id:             # Plugin id(s) to be used in place of the default for the task type.
           - plugin_override1
           - plugin_override2
         missing_plugin_behavior: 1 # Behavior when no specified plugin_id has an associated handler. 0 : FAIL , 1: DEFAULT

::

 flytectl update plugin-override --attrFile po.yaml

Updating plugin override for project and domain and workflow combination. This will take precedence over any other
plugin overrides defined at project domain level.
Update the plugin overrides for workflow core.control_flow.run_merge_sort.merge_sort in flytectldemo , development domain

.. code-block:: yaml

    domain: development
    project: flytectldemo
    workflow: core.control_flow.run_merge_sort.merge_sort
    overrides:
       - task_type: python_task # Task type for which to apply plugin implementation overrides
         plugin_id:             # Plugin id(s) to be used in place of the default for the task type.
           - plugin_override1
           - plugin_override2
         missing_plugin_behavior: 1 # Behavior when no specified plugin_id has an associated handler. 0 : FAIL , 1: DEFAULT

::

 flytectl update plugin-override --attrFile po.yaml

Usage



::

  flytectl update plugin-override [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for updating attribute for the resource type.
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

* :doc:`flytectl_update` 	 - Used for updating flyte resources eg: project.

