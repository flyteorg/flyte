.. _flytectl_sandbox_teardown:

flytectl sandbox teardown
-------------------------

Teardown will cleanup the sandbox environment

Synopsis
~~~~~~~~



Teardown will remove docker container and all the flyte config 
::

 bin/flytectl sandbox teardown 
	

Usage


::

  flytectl sandbox teardown [flags]

Options
~~~~~~~

::

  -h, --help   help for teardown

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

