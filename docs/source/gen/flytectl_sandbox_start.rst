.. _flytectl_sandbox_start:

flytectl sandbox start
----------------------

Start the flyte sandbox cluster

Synopsis
~~~~~~~~



The Flyte Sandbox is a fully standalone minimal environment for running Flyte. provides a simplified way of running flyte-sandbox as a single Docker container running locally.  

Start sandbox cluster without any source code
::

 bin/flytectl sandbox start
	
Mount your source code repository inside sandbox 
::

 bin/flytectl sandbox start --source=$HOME/flyteorg/flytesnacks 

Usage
	

::

  flytectl sandbox start [flags]

Options
~~~~~~~

::

  -h, --help            help for start
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

