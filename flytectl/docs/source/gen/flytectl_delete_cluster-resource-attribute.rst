.. _flytectl_delete_cluster-resource-attribute:

flytectl delete cluster-resource-attribute
------------------------------------------

Deletes matchable resources of cluster attributes

Synopsis
~~~~~~~~



Deletes cluster resource attributes for given project and domain combination or additionally with workflow name.

Deletes cluster resource attribute for project and domain
Here the command delete cluster resource attributes for  project flytectldemo and development domain.
::

 flytectl delete cluster-resource-attribute -p flytectldemo -d development 


Deletes cluster resource attribute using config file which was used for creating it.
Here the command deletes cluster resource attributes from the config file cra.yaml
Attributes are optional in the file as they are unread during the delete command but can be kept as the same file can be used for get, update or delete 
eg:  content of cra.yaml which will use the project domain and workflow name for deleting the resource

::

 flytectl delete cluster-resource-attribute --attrFile cra.yaml


.. code-block:: yaml
	
    domain: development
    project: flytectldemo
    attributes:
      foo: "bar"
      buzz: "lightyear"

Deletes cluster resource attribute for a workflow
Here the command deletes cluster resource attributes for a workflow core.control_flow.run_merge_sort.merge_sort

::

 flytectl delete cluster-resource-attribute -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

Usage


::

  flytectl delete cluster-resource-attribute [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for delete attribute for the resource type.
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

* :doc:`flytectl_delete` 	 - Used for terminating/deleting various flyte resources including tasks/workflows/launchplans/executions/project.

