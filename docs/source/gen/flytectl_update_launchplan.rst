.. _flytectl_update_launchplan:

flytectl update launchplan
--------------------------

Updates launch plan metadata

Synopsis
~~~~~~~~



Following command updates the description on the launchplan.
::

 flytectl update launchplan -p flytectldemo -d development  core.advanced.run_merge_sort.merge_sort --description "Mergesort example"

Archiving launchplan named entity is not supported and would throw an error.
::

 flytectl update launchplan -p flytectldemo -d development  core.advanced.run_merge_sort.merge_sort --archive

Activating launchplan named entity would be a noop.
::

 flytectl update launchplan -p flytectldemo -d development  core.advanced.run_merge_sort.merge_sort --activate

Usage


::

  flytectl update launchplan [flags]

Options
~~~~~~~

::

      --activate             Activates the named entity specified as argument.
      --archive              Archives the named entity specified as argument.
      --description string   description of the namedentity.
  -h, --help                 help for launchplan

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

