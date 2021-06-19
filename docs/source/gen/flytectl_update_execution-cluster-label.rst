.. _flytectl_update_execution-cluster-label:

flytectl update execution-cluster-label
---------------------------------------

Updates matchable resources of execution cluster label

Synopsis
~~~~~~~~



Updates execution cluster label for given project and domain combination or additionally with workflow name.

Updating to the execution cluster label is only available from a generated file. See the get section for generating this file.
Here the command updates takes the input for execution cluster label from the config file ecl.yaml
eg:  content of ecl.yaml

.. code-block:: yaml

    domain: development
    project: flytectldemo
    value: foo

::

 flytectl update execution-cluster-label --attrFile ecl.yaml

Updating execution cluster label for project and domain and workflow combination. This will take precedence over any other
execution cluster label defined at project domain level.
Update the execution cluster label for workflow core.control_flow.run_merge_sort.merge_sort in flytectldemo, development domain

.. code-block:: yaml

    domain: development
    project: flytectldemo
    workflow: core.control_flow.run_merge_sort.merge_sort
    value: foo

::

 flytectl update execution-cluster-label --attrFile ecl.yaml

Usage



::

  flytectl update execution-cluster-label [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for updating attribute for the resource type.
  -h, --help              help for execution-cluster-label

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

