.. _flytectl_config:

flytectl config
---------------

Runs various config commands, look at the help of this command to get a list of available commands..

Synopsis
~~~~~~~~


Runs various config commands, look at the help of this command to get a list of available commands..

Options
~~~~~~~

::

      --file stringArray   Passes the config file to load.
                           If empty, it'll first search for the config file path then, if found, will load config from there.
  -h, --help               help for config

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
* :doc:`flytectl_config_discover` 	 - Searches for a config in one of the default search paths.
* :doc:`flytectl_config_validate` 	 - Validates the loaded config.

