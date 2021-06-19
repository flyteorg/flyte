.. _flytectl_get_task-resource-attribute:

flytectl get task-resource-attribute
------------------------------------

Gets matchable resources of task attributes

Synopsis
~~~~~~~~



Retrieves task  resource attributes for given project,domain combination or additionally with workflow name.

Retrieves task resource attribute for project and domain
Here the command get task resource attributes for  project flytectldemo and development domain.
::

 flytectl get task-resource-attribute -p flytectldemo -d development 

eg : output from the command

.. code-block:: json

 {"project":"flytectldemo","domain":"development","workflow":"","defaults":{"cpu":"1","memory":"150Mi"},"limits":{"cpu":"2","memory":"450Mi"}}

Retrieves task resource attribute for project and domain and workflow
Here the command get task resource attributes for  project flytectldemo, development domain and workflow core.control_flow.run_merge_sort.merge_sort
::

 flytectl get task-resource-attribute -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

eg : output from the command

.. code-block:: json

 {"project":"flytectldemo","domain":"development","workflow":"core.control_flow.run_merge_sort.merge_sort","defaults":{"cpu":"1","memory":"150Mi"},"limits":{"cpu":"2","memory":"450Mi"}}


Writing the task resource attribute to a file. If there are no task resource attributes a file would be written with basic data populated.
Here the command gets task resource attributes and writes the config file to tra.yaml
eg:  content of tra.yaml

::

 flytectl get task-resource-attribute --attrFile tra.yaml


.. code-block:: yaml

    domain: development
    project: flytectldemo
    defaults:
      cpu: "1"
      memory: "150Mi"
    limits:
      cpu: "2"
      memory: "450Mi"

Usage


::

  flytectl get task-resource-attribute [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for generating attribute for the resource type.
  -h, --help              help for task-resource-attribute

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

