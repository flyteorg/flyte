.. _flytectl_update_workflow-execution-config:

flytectl update workflow-execution-config
-----------------------------------------

Updates matchable resources of workflow execution config

Synopsis
~~~~~~~~



Updates workflow execution config for given project and domain combination or additionally with workflow name.

Updating the workflow execution config is only available from a generated file. See the get section for generating this file.
Also this will completely overwrite any existing custom project and domain and workflow combination execution config.
Would be preferable to do get and generate an config file if there is an existing execution config already set and then update it to have new values
Refer to get workflow-execution-config section on how to generate this file
Here the command updates takes the input for workflow execution config from the config file wec.yaml
eg:  content of wec.yaml

.. code-block:: yaml

    domain: development
    project: flytectldemo
    max_parallelism: 5

::

 flytectl update workflow-execution-config --attrFile wec.yaml

Updating workflow execution config for project and domain and workflow combination. This will take precedence over any other
execution config defined at project domain level.
Update the workflow execution config for workflow core.control_flow.run_merge_sort.merge_sort in flytectldemo, development domain

.. code-block:: yaml

    domain: development
    project: flytectldemo
    workflow: core.control_flow.run_merge_sort.merge_sort
    max_parallelism: 5

::

 flytectl update workflow-execution-config --attrFile wec.yaml

Usage



::

  flytectl update workflow-execution-config [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for updating attribute for the resource type.
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

* :doc:`flytectl_update` 	 - Used for updating flyte resources eg: project.

