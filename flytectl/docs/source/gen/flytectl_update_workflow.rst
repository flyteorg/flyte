.. _flytectl_update_workflow:

flytectl update workflow
------------------------

Updates workflow metadata

Synopsis
~~~~~~~~



Following command updates the description on the workflow.
::

 flytectl update workflow -p flytectldemo -d development core.advanced.run_merge_sort.merge_sort --description "Mergesort workflow example"

Archiving workflow named entity would cause this to disapper from flyteconsole UI.
::

 flytectl update workflow -p flytectldemo -d development  core.advanced.run_merge_sort.merge_sort --archive

Activating workflow named entity
::

 flytectl update workflow -p flytectldemo -d development  core.advanced.run_merge_sort.merge_sort --activate

Usage


::

  flytectl update workflow [flags]

Options
~~~~~~~

::

      --activate             activate the named entity.
      --archive              archive named entity.
      --description string   description of the named entity.
  -h, --help                 help for workflow

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

