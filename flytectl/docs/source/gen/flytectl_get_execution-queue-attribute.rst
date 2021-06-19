.. _flytectl_get_execution-queue-attribute:

flytectl get execution-queue-attribute
--------------------------------------

Gets matchable resources of execution queue attributes

Synopsis
~~~~~~~~



Retrieves execution queue attributes for given project and domain combination or additionally with workflow name.

Retrieves execution queue attribute for project and domain
Here the command get execution queue attributes for  project flytectldemo and development domain.
::

 flytectl get execution-queue-attribute -p flytectldemo -d development 

eg : output from the command

.. code-block:: json

 {"project":"flytectldemo","domain":"development","tags":["foo", "bar"]}

Retrieves execution queue attribute for project and domain and workflow
Here the command get execution queue attributes for  project flytectldemo, development domain and workflow core.control_flow.run_merge_sort.merge_sort
::

 flytectl get execution-queue-attribute -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

eg : output from the command

.. code-block:: json

 {"project":"flytectldemo","domain":"development","workflow":"core.control_flow.run_merge_sort.merge_sort","tags":["foo", "bar"]}

Writing the execution queue attribute to a file. If there are no execution queue attributes, command would return an error.
Here the command gets execution queue attributes and writes the config file to era.yaml
eg:  content of era.yaml

::

 flytectl get execution-queue-attribute --attrFile era.yaml


.. code-block:: yaml

    domain: development
    project: flytectldemo
    tags:
      - foo
      - bar
      - buzz
      - lightyear

Usage


::

  flytectl get execution-queue-attribute [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for generating attribute for the resource type.
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

* :doc:`flytectl_get` 	 - Used for fetching various flyte resources including tasks/workflows/launchplans/executions/project.

