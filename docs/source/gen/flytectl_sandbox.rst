.. _flytectl_sandbox:

flytectl sandbox
----------------

Used for sandbox interactions like start/teardown/status/exec.

Synopsis
~~~~~~~~



The Flyte Sandbox is a fully standalone minimal environment for running Flyte. provides a simplified way of running flyte-sandbox as a single Docker container running locally.
	
Create sandbox cluster.
::

 bin/flytectl sandbox start 
	
	
Remove sandbox cluster.
::

 bin/flytectl sandbox teardown 	
	

Check status of sandbox container.
::

 bin/flytectl sandbox status 	
	
Execute command inside sandbox container.
::

 bin/flytectl sandbox exec -- pwd 	


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
* :doc:`flytectl_sandbox_exec` 	 - Execute non-interactive command inside the sandbox container
* :doc:`flytectl_sandbox_start` 	 - Start the flyte sandbox cluster
* :doc:`flytectl_sandbox_status` 	 - Get the status of the sandbox environment.
* :doc:`flytectl_sandbox_teardown` 	 - Teardown will cleanup the sandbox environment

