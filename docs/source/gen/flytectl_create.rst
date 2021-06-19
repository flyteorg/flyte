.. _flytectl_create:

flytectl create
---------------

Used for creating various flyte resources including tasks/workflows/launchplans/executions/project.

Synopsis
~~~~~~~~



Example create.
::

 bin/flytectl create project --file project.yaml 


Options
~~~~~~~

::

  -h, --help   help for create

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  -c, --config string    config file (default is $HOME/.flyte/config.yaml)
  -d, --domain string    Specifies the Flyte project's domain.
  -o, --output string    Specifies the output type - supported formats [TABLE JSON YAML DOT DOTURL]. NOTE: dot, doturl are only supported for Workflow (default "TABLE")
  -p, --project string   Specifies the Flyte project.

SEE ALSO
~~~~~~~~

* :doc:`flytectl` 	 - flyetcl CLI tool
* :doc:`flytectl_create_execution` 	 - Create execution resources
* :doc:`flytectl_create_project` 	 - Create project resources

