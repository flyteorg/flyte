.. _flytectl_register_examples:

flytectl register examples
--------------------------

Registers flytesnack example

Synopsis
~~~~~~~~



Registers all latest flytesnacks example
::

 bin/flytectl register examples  -d development  -p flytesnacks


Usage


::

  flytectl register examples [flags]

Options
~~~~~~~

::

  -a, --archive                       pass in archive file either an http link or local path.
  -i, --assumableIamRole string        Custom assumable iam auth role to register launch plans with.
      --continueOnError               continue on error when registering files.
  -h, --help                          help for examples
  -k, --k8ServiceAccount string        custom kubernetes service account auth role to register launch plans with.
  -l, --outputLocationPrefix string    custom output location prefix for offloaded types (files/schemas).
      --sourceUploadPath string        Location for source code in storage.
  -v, --version string                version of the entity to be registered with flyte. (default "v1")

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  -c, --config string    config file (default is $HOME/.flyte/config.yaml)
  -d, --domain string    Specifies the Flyte project's domain.
  -o, --output string    Specifies the output type - supported formats [TABLE JSON YAML DOT DOTURL]. NOTE: dot, doturl are only supported for Workflow (default "TABLE")
  -p, --project string   Specifies the Flyte project.

SEE ALSO
~~~~~~~~

* :doc:`flytectl_register` 	 - Registers tasks/workflows/launchplans from list of generated serialized files.

