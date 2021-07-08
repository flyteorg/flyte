.. _flytectl_sandbox_status:

flytectl sandbox status
-----------------------

Get the status of the sandbox environment.

Synopsis
~~~~~~~~



Status will retrieve the status of the Sandbox environment. Currently FlyteSandbox runs as a local docker container.
This will return the docker status for this container

Usage
::

 bin/flytectl sandbox status 



::

  flytectl sandbox status [flags]

Options
~~~~~~~

::

  -h, --help   help for status

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

