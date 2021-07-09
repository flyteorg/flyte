.. _flytectl_update_task:

flytectl update task
--------------------

Updates task metadata

Synopsis
~~~~~~~~



Following command updates the description on the task.
::

 flytectl update  task -d development -p flytectldemo core.advanced.run_merge_sort.merge --description "Merge sort example"

Archiving task named entity is not supported and would throw an error.
::

 flytectl update  task -d development -p flytectldemo core.advanced.run_merge_sort.merge --archive

Activating task named entity would be a noop as archiving is not possible.
::

 flytectl update  task -d development -p flytectldemo core.advanced.run_merge_sort.merge --activate

Usage


::

  flytectl update task [flags]

Options
~~~~~~~

::

      --activate             activate the named entity.
      --archive              archive named entity.
      --description string   description of the named entity.
  -h, --help                 help for task

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

