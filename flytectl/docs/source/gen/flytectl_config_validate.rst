.. _flytectl_config_validate:

flytectl config validate
------------------------

Validates the loaded config.

Synopsis
~~~~~~~~


Validates the loaded config.

::

  flytectl config validate [flags]

Options
~~~~~~~

::

  -h, --help     help for validate
      --strict   Validates that all keys in loaded config
                 map to already registered sections.

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  -c, --config string      config file (default is $HOME/.flyte/config.yaml)
  -d, --domain string      Specifies the Flyte project's domain.
      --file stringArray   Passes the config file to load.
                           If empty, it'll first search for the config file path then, if found, will load config from there.
  -o, --output string      Specifies the output type - supported formats [TABLE JSON YAML DOT DOTURL]. NOTE: dot, doturl are only supported for Workflow (default "TABLE")
  -p, --project string     Specifies the Flyte project.

SEE ALSO
~~~~~~~~

* :doc:`flytectl_config` 	 - Runs various config commands, look at the help of this command to get a list of available commands..

