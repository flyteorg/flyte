.. _flytectl_register:

flytectl register
-----------------

Registers tasks/workflows/launchplans from list of generated serialized files.

Synopsis
~~~~~~~~



Takes input files as serialized versions of the tasks/workflows/launchplans and registers them with flyteadmin.
Currently these input files are protobuf files generated as output from flytekit serialize.
Project & Domain are mandatory fields to be passed for registration and an optional version which defaults to v1
If the entities are already registered with flyte for the same version then registration would fail.


Options
~~~~~~~

::

  -h, --help   help for register

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
* :doc:`flytectl_register_examples` 	 - Registers flytesnack example
* :doc:`flytectl_register_files` 	 - Registers file resources

