.. _flytectl_update_project:

flytectl update project
-----------------------

Updates project resources

Synopsis
~~~~~~~~



Updates the project according the flags passed. Allows you to archive or activate a project.
Activates project named flytesnacks.
::

 bin/flytectl update project -p flytesnacks --activateProject

Archives project named flytesnacks.

::

 bin/flytectl update project -p flytesnacks --archiveProject

Activates project named flytesnacks using short option -t.
::

 bin/flytectl update project -p flytesnacks -t

Archives project named flytesnacks using short option -a.

::

 bin/flytectl update project flytesnacks -a

Incorrect usage when passing both archive and activate.

::

 bin/flytectl update project flytesnacks -a -t

Incorrect usage when passing unknown-project.

::

 bin/flytectl update project unknown-project -a

Incorrect usage when passing valid project using -p option.

::

 bin/flytectl update project unknown-project -a -p known-project

Usage


::

  flytectl update project [flags]

Options
~~~~~~~

::

  -t, --activateProject   Activates the project specified as argument.
  -a, --archiveProject    Archives the project specified as argument.
  -h, --help              help for project

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

