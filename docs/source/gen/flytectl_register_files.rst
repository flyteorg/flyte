.. _flytectl_register_files:

flytectl register files
-----------------------

Registers file resources

Synopsis
~~~~~~~~



Registers all the serialized protobuf files including tasks, workflows and launchplans with default v1 version.
If there are already registered entities with v1 version then the command will fail immediately on the first such encounter.
::

 bin/flytectl register file  _pb_output/* -d development  -p flytesnacks

Using archive file.Currently supported are .tgz and .tar extension files and can be local or remote file served through http/https.
Use --archive flag.

::

 bin/flytectl register files  http://localhost:8080/_pb_output.tar -d development  -p flytesnacks --archive

Using  local tgz file.

::

 bin/flytectl register files  _pb_output.tgz -d development  -p flytesnacks --archive

If you want to continue executing registration on other files ignoring the errors including version conflicts then pass in
the continueOnError flag.

::

 bin/flytectl register file  _pb_output/* -d development  -p flytesnacks --continueOnError

Using short format of continueOnError flag
::

 bin/flytectl register file  _pb_output/* -d development  -p flytesnacks --continueOnError

Overriding the default version v1 using version string.
::

 bin/flytectl register file  _pb_output/* -d development  -p flytesnacks -v v2

Change the o/p format has not effect on registration. The O/p is currently available only in table format.

::

 bin/flytectl register file  _pb_output/* -d development  -p flytesnacks --continueOnError -o yaml

Override IamRole during registration.

::

 bin/flytectl register file  _pb_output/* -d development  -p flytesnacks --continueOnError -v v2 -i "arn:aws:iam::123456789:role/dummy"

Override Kubernetes service account during registration.

::

 bin/flytectl register file  _pb_output/* -d development  -p flytesnacks --continueOnError -v v2 -k "kubernetes-service-account"

Override Output location prefix during registration.

::

 bin/flytectl register file  _pb_output/* -d development  -p flytesnacks --continueOnError -v v2 -l "s3://dummy/prefix"

Usage


::

  flytectl register files [flags]

Options
~~~~~~~

::

  -a, --archive                       pass in archive file either an http link or local path.
  -i, --assumableIamRole string        Custom assumable iam auth role to register launch plans with.
      --continueOnError               continue on error when registering files.
  -h, --help                          help for files
  -k, --k8ServiceAccount string        custom kubernetes service account auth role to register launch plans with.
  -l, --outputLocationPrefix string    custom output location prefix for offloaded types (files/schemas).
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

