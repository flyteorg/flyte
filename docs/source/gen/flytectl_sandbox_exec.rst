.. _flytectl_sandbox_exec:

flytectl sandbox exec
---------------------

Execute any command in sandbox

Synopsis
~~~~~~~~



Execute command will Will run non-interactive commands and return immediately with the output.

::
 bin/flytectl sandbox exec -- ls -al 

Usage

::

  flytectl sandbox exec [flags]

Options
~~~~~~~

::

  -h, --help            help for exec
      --source string    Path of your source code

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  -c, --config string    config file (default is $HOME/.flyte/config.yaml)
  -d, --domain string    Specifies the Flyte project's domain.
  -o, --output string    Specifies the output type - supported formats [TABLE JSON YAML DOT DOTURL]. NOTE: dot, doturl are only supported for Workflow (default "TABLE")
  -p, --project string   Specifies the Flyte project.

SEE ALSO
~~~~~~~~

* :doc:`flytectl_sandbox` 	 - Used for testing flyte sandbox.

