.. _flytectl_sandbox_start:

flytectl sandbox start
----------------------

Start the flyte sandbox

Synopsis
~~~~~~~~



Start will run the flyte sandbox cluster inside a docker container and setup the config that is required 
::

 bin/flytectl sandbox start
	
Mount your flytesnacks repository code inside sandbox 
::

 bin/flytectl sandbox start --sourcesPath=$HOME/flyteorg/flytesnacks
Usage
	

::

  flytectl sandbox start [flags]

Options
~~~~~~~

::

      --sourcesPath string   Path to your source code path where flyte workflows and tasks are.
  -h, --help                 help for start

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

