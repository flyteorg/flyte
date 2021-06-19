.. _flytectl_update_execution-queue-attribute:

flytectl update execution-queue-attribute
-----------------------------------------

Updates matchable resources of execution queue attributes

Synopsis
~~~~~~~~



Updates execution queue attributes for given project and domain combination or additionally with workflow name.

Updating to the execution queue attribute is only available from a generated file. See the get section for generating this file.
Also this will completely overwrite any existing custom project and domain and workflow combination attributes.
Would be preferable to do get and generate an attribute file if there is an existing attribute already set and then update it to have new values
Refer to get execution-queue-attribute section on how to generate this file
Here the command updates takes the input for execution queue attributes from the config file era.yaml
eg:  content of era.yaml

.. code-block:: yaml

    domain: development
    project: flytectldemo
    tags:
      - foo
      - bar
      - buzz
      - lightyear

::

 flytectl update execution-queue-attribute --attrFile era.yaml

Updating execution queue attribute for project and domain and workflow combination. This will take precedence over any other
execution queue attribute defined at project domain level.
Update the execution queue attributes for workflow core.control_flow.run_merge_sort.merge_sort in flytectldemo, development domain

.. code-block:: yaml

    domain: development
    project: flytectldemo
    workflow: core.control_flow.run_merge_sort.merge_sort
    tags:
      - foo
      - bar
      - buzz
      - lightyear

::

 flytectl update execution-queue-attribute --attrFile era.yaml

Usage



::

  flytectl update execution-queue-attribute [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for updating attribute for the resource type.
  -h, --help              help for execution-queue-attribute

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

