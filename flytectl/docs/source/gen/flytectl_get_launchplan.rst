.. _flytectl_get_launchplan:

flytectl get launchplan
-----------------------

Gets launch plan resources

Synopsis
~~~~~~~~



Retrieves all the launch plans within project and domain.(launchplan,launchplans can be used interchangeably in these commands)
::

 flytectl get launchplan -p flytesnacks -d development

Retrieves launch plan by name within project and domain.

::

 flytectl get launchplan -p flytesnacks -d development core.basic.lp.go_greet


Retrieves latest version of task by name within project and domain.

::

 flytectl get launchplan -p flytesnacks -d development  core.basic.lp.go_greet --latest

Retrieves particular version of launchplan by name within project and domain.

::

 flytectl get launchplan -p flytesnacks -d development  core.basic.lp.go_greet --version v2

Retrieves launchplan by filters.
::

 Not yet implemented

Retrieves all the launchplan within project and domain in yaml format.

::

 flytectl get launchplan -p flytesnacks -d development -o yaml

Retrieves all the launchplan within project and domain in json format

::

 flytectl get launchplan -p flytesnacks -d development -o json

Retrieves a launch plans within project and domain for a version and generate the execution spec file for it to be used for launching the execution using create execution.

::

 flytectl get launchplan -d development -p flytectldemo core.advanced.run_merge_sort.merge_sort --execFile execution_spec.yam

The generated file would look similar to this

.. code-block:: yaml

	 iamRoleARN: ""
	 inputs:
	   numbers:
	   - 0
	   numbers_count: 0
	   run_local_at_count: 10
	 kubeServiceAcct: ""
	 targetDomain: ""
	 targetProject: ""
	 version: v3
	 workflow: core.advanced.run_merge_sort.merge_sort

Check the create execution section on how to launch one using the generated file.

Usage


::

  flytectl get launchplan [flags]

Options
~~~~~~~

::

      --execFile string   execution file name to be used for generating execution spec of a single launchplan.
  -h, --help              help for launchplan
      --latest            flag to indicate to fetch the latest version, version flag will be ignored in this case
      --version string    version of the launchplan to be fetched.

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --admin.authorizationHeader string           Custom metadata header to pass JWT
      --admin.authorizationServerUrl string        This is the URL to your IdP's authorization server. It'll default to Endpoint
      --admin.clientId string                      Client ID (default "flytepropeller")
      --admin.clientSecretLocation string          File containing the client secret (default "/etc/secrets/client_secret")
      --admin.endpoint string                      For admin types,  specify where the uri of the service is located.
      --admin.insecure                             Use insecure connection.
      --admin.maxBackoffDelay string               Max delay for grpc backoff (default "8s")
      --admin.maxRetries int                       Max number of gRPC retries (default 4)
      --admin.perRetryTimeout string               gRPC per retry timeout (default "15s")
      --admin.scopes strings                       List of scopes to request
      --admin.tokenUrl string                      OPTIONAL: Your IdP's token endpoint. It'll be discovered from flyte admin's OAuth Metadata endpoint if not provided.
      --admin.useAuth                              Deprecated: Auth will be enabled/disabled based on admin's dynamically discovered information.
      --adminutils.batchSize int                   Maximum number of records to retrieve per call. (default 100)
      --adminutils.maxRecords int                  Maximum number of records to retrieve. (default 500)
      --config string                              config file (default is $HOME/.flyte/config.yaml)
  -d, --domain string                              Specifies the Flyte project's domain.
      --logger.formatter.type string               Sets logging format type. (default "json")
      --logger.level int                           Sets the minimum logging level. (default 4)
      --logger.mute                                Mutes all logs regardless of severity. Intended for benchmarks/tests only.
      --logger.show-source                         Includes source code location in logs.
  -o, --output string                              Specifies the output type - supported formats [TABLE JSON YAML] (default "TABLE")
  -p, --project string                             Specifies the Flyte project.
      --root.domain string                         Specified the domain to work on.
      --root.output string                         Specified the output type.
      --root.project string                        Specifies the project to work on.
      --storage.cache.max_size_mbs int             Maximum size of the cache where the Blob store data is cached in-memory. If not specified or set to 0,  cache is not used
      --storage.cache.target_gc_percent int        Sets the garbage collection target percentage.
      --storage.connection.access-key string       Access key to use. Only required when authtype is set to accesskey.
      --storage.connection.auth-type string        Auth Type to use [iam, accesskey]. (default "iam")
      --storage.connection.disable-ssl             Disables SSL connection. Should only be used for development.
      --storage.connection.endpoint string         URL for storage client to connect to.
      --storage.connection.region string           Region to connect to. (default "us-east-1")
      --storage.connection.secret-key string       Secret to use when accesskey is set.
      --storage.container string                   Initial container to create -if it doesn't exist-.'
      --storage.defaultHttpClient.timeout string   Sets time out on the http client. (default "0s")
      --storage.enable-multicontainer              If this is true,  then the container argument is overlooked and redundant. This config will automatically open new connections to new containers/buckets as they are encountered
      --storage.limits.maxDownloadMBs int          Maximum allowed download size (in MBs) per call. (default 2)
      --storage.type string                        Sets the type of storage to configure [s3/minio/local/mem/stow]. (default "s3")

SEE ALSO
~~~~~~~~

* :doc:`flytectl_get` 	 - Used for fetching various flyte resources including tasks/workflows/launchplans/executions/project.

