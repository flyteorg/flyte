.. _flytectl_sandbox_exec:

flytectl sandbox exec
---------------------

Execute non-interactive command inside the sandbox container

Synopsis
~~~~~~~~



Execute command will run non-interactive command inside the sandbox container and return immediately with the output.By default flytectl exec in /root directory inside the sandbox container

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

* :doc:`flytectl_sandbox` 	 - Used for sandbox interactions like start/teardown/status/exec.

