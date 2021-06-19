.. _flytectl_update_cluster-resource-attribute:

flytectl update cluster-resource-attribute
------------------------------------------

Updates matchable resources of cluster attributes

Synopsis
~~~~~~~~



Updates cluster resource attributes for given project and domain combination or additionally with workflow name.

Updating to the cluster resource attribute is only available from a generated file. See the get section for generating this file.
Here the command updates takes the input for cluster resource attributes from the config file cra.yaml
eg:  content of cra.yaml

.. code-block:: yaml

    domain: development
    project: flytectldemo
    attributes:
      foo: "bar"
      buzz: "lightyear"

::

 flytectl update cluster-resource-attribute --attrFile cra.yaml

Updating cluster resource attribute for project and domain and workflow combination. This will take precedence over any other
resource attribute defined at project domain level.
Also this will completely overwrite any existing custom project and domain and workflow combination attributes.
Would be preferable to do get and generate an attribute file if there is an existing attribute already set and then update it to have new values
Refer to get cluster-resource-attribute section on how to generate this file
Update the cluster resource attributes for workflow core.control_flow.run_merge_sort.merge_sort in flytectldemo, development domain

.. code-block:: yaml

    domain: development
    project: flytectldemo
    workflow: core.control_flow.run_merge_sort.merge_sort
    attributes:
      foo: "bar"
      buzz: "lightyear"

::

 flytectl update cluster-resource-attribute --attrFile cra.yaml

Usage



::

  flytectl update cluster-resource-attribute [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for updating attribute for the resource type.
  -h, --help              help for cluster-resource-attribute

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

