.. _flytectl_delete_execution-queue-attribute:

flytectl delete execution-queue-attribute
-----------------------------------------

Deletes matchable resources of execution queue attributes

Synopsis
~~~~~~~~



Deletes execution queue attributes for given project and domain combination or additionally with workflow name.

Deletes execution queue attribute for project and domain
Here the command delete execution queue attributes for project flytectldemo and development domain.
::

 flytectl delete execution-queue-attribute -p flytectldemo -d development 


Deletes execution queue attribute using config file which was used for creating it.
Here the command deletes execution queue attributes from the config file era.yaml
Tags are optional in the file as they are unread during the delete command but can be kept as the same file can be used for get, update or delete 
eg:  content of era.yaml which will use the project domain and workflow name for deleting the resource

::

 flytectl delete execution-queue-attribute --attrFile era.yaml


.. code-block:: yaml

    domain: development
    project: flytectldemo
    tags:
      - foo
      - bar
      - buzz
      - lightyear

Deletes execution queue attribute for a workflow
Here the command deletes the execution queue attributes for a workflow core.control_flow.run_merge_sort.merge_sort

::

 flytectl delete execution-queue-attribute -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

Usage


::

  flytectl delete execution-queue-attribute [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for delete attribute for the resource type.
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

* :doc:`flytectl_delete` 	 - Used for terminating/deleting various flyte resources including tasks/workflows/launchplans/executions/project.

