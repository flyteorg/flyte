.. _flytectl_delete_execution-cluster-label:

flytectl delete execution-cluster-label
---------------------------------------

Deletes matchable resources of execution cluster label

Synopsis
~~~~~~~~



Deletes execution cluster label for given project and domain combination or additionally with workflow name.

Deletes execution cluster label for project and domain
Here the command delete execution cluster label for project flytectldemo and development domain.
::

 flytectl delete execution-cluster-label -p flytectldemo -d development 


Deletes execution cluster label using config file which was used for creating it.
Here the command deletes execution cluster label from the config file ecl.yaml
Value is optional in the file as its unread during the delete command but can be kept as the same file can be used for get, update or delete 
eg:  content of ecl.yaml which will use the project domain and workflow name for deleting the resource

::

 flytectl delete execution-cluster-label --attrFile ecl.yaml


.. code-block:: yaml
	
    domain: development
    project: flytectldemo
    value: foo

Deletes execution cluster label for a workflow
Here the command deletes execution cluster label for a workflow core.control_flow.run_merge_sort.merge_sort

::

 flytectl delete execution-cluster-label -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

Usage


::

  flytectl delete execution-cluster-label [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for delete attribute for the resource type.
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

* :doc:`flytectl_delete` 	 - Used for terminating/deleting various flyte resources including tasks/workflows/launchplans/executions/project.

