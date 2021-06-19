.. _flytectl_sandbox:

flytectl sandbox
----------------

Used for testing flyte sandbox.

Synopsis
~~~~~~~~



Example Create sandbox cluster.
::

 bin/flytectl sandbox start 
	
	
Example Remove sandbox cluster.
::

 bin/flytectl sandbox teardown 	


Options
~~~~~~~

::

  -h, --help   help for sandbox

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
* :doc:`flytectl_sandbox_start` 	 - Start the flyte sandbox
* :doc:`flytectl_sandbox_teardown` 	 - Teardown will cleanup the sandbox environment

