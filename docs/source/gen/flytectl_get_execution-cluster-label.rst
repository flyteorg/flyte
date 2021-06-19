.. _flytectl_get_execution-cluster-label:

flytectl get execution-cluster-label
------------------------------------

Gets matchable resources of execution cluster label

Synopsis
~~~~~~~~



Retrieves execution cluster label for given project and domain combination or additionally with workflow name.

Retrieves execution cluster label for project and domain
Here the command get execution cluster label for project flytectldemo and development domain.
::

 flytectl get execution-cluster-label -p flytectldemo -d development 

eg : output from the command

.. code-block:: json

 {"project":"flytectldemo","domain":"development","value":"foo"}

Retrieves execution cluster label for project and domain and workflow
Here the command get execution cluster label for  project flytectldemo, development domain and workflow core.control_flow.run_merge_sort.merge_sort
::

 flytectl get execution-cluster-label -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

eg : output from the command

.. code-block:: json

 {"project":"flytectldemo","domain":"development","workflow":"core.control_flow.run_merge_sort.merge_sort","value":"foo"}

Writing the execution cluster label to a file. If there are no execution cluster label, command would return an error.
Here the command gets execution cluster label and writes the config file to ecl.yaml
eg:  content of ecl.yaml

::

 flytectl get execution-cluster-label --attrFile ecl.yaml


.. code-block:: yaml

    domain: development
    project: flytectldemo
    value: foo

Usage


::

  flytectl get execution-cluster-label [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for generating attribute for the resource type.
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

* :doc:`flytectl_get` 	 - Used for fetching various flyte resources including tasks/workflows/launchplans/executions/project.

