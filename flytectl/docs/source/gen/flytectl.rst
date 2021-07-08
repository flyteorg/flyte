.. _flytectl:

flytectl
--------

flyetcl CLI tool

Synopsis
~~~~~~~~


flytectl is CLI tool written in go to interact with flyteadmin service

Options
~~~~~~~

::

  -c, --config string    config file (default is $HOME/.flyte/config.yaml)
  -d, --domain string    Specifies the Flyte project's domain.
  -h, --help             help for flytectl
  -o, --output string    Specifies the output type - supported formats [TABLE JSON YAML DOT DOTURL]. NOTE: dot, doturl are only supported for Workflow (default "TABLE")
  -p, --project string   Specifies the Flyte project.

SEE ALSO
~~~~~~~~

* :doc:`flytectl_completion` 	 - Generate completion script
* :doc:`flytectl_config` 	 - Runs various config commands, look at the help of this command to get a list of available commands..
* :doc:`flytectl_create` 	 - Used for creating various flyte resources including tasks/workflows/launchplans/executions/project.
* :doc:`flytectl_delete` 	 - Used for terminating/deleting various flyte resources including tasks/workflows/launchplans/executions/project.
* :doc:`flytectl_get` 	 - Used for fetching various flyte resources including tasks/workflows/launchplans/executions/project.
* :doc:`flytectl_register` 	 - Registers tasks/workflows/launchplans from list of generated serialized files.
* :doc:`flytectl_sandbox` 	 - Used for sandbox interactions like start/teardown/status/exec.
* :doc:`flytectl_update` 	 - Used for updating flyte resources eg: project.
* :doc:`flytectl_version` 	 - Used for fetching flyte version

