.. _flytectl_get_cluster-resource-attribute:

flytectl get cluster-resource-attribute
---------------------------------------

Gets matchable resources of cluster resource attributes

Synopsis
~~~~~~~~



Retrieves cluster resource attributes for given project and domain combination or additionally with workflow name.

Retrieves cluster resource attribute for project and domain
Here the command get cluster resource attributes for  project flytectldemo and development domain.
::

 flytectl get cluster-resource-attribute -p flytectldemo -d development 

eg : output from the command

.. code-block:: json

 {"project":"flytectldemo","domain":"development","attributes":{"buzz":"lightyear","foo":"bar"}}

Retrieves cluster resource attribute for project and domain and workflow
Here the command get cluster resource attributes for  project flytectldemo, development domain and workflow core.control_flow.run_merge_sort.merge_sort
::

 flytectl get cluster-resource-attribute -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

eg : output from the command

.. code-block:: json

 {"project":"flytectldemo","domain":"development","workflow":"core.control_flow.run_merge_sort.merge_sort","attributes":{"buzz":"lightyear","foo":"bar"}}

Writing the cluster resource attribute to a file. If there are no cluster resource attributes , command would return an error.
Here the command gets task resource attributes and writes the config file to cra.yaml
eg:  content of cra.yaml

::

 flytectl get task-resource-attribute --attrFile cra.yaml


.. code-block:: yaml

    domain: development
    project: flytectldemo
    attributes:
      foo: "bar"
      buzz: "lightyear"

Usage


::

  flytectl get cluster-resource-attribute [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for generating attribute for the resource type.
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

* :doc:`flytectl_get` 	 - Used for fetching various flyte resources including tasks/workflows/launchplans/executions/project.

