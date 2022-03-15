.. _flytectl_get_launchplan:

flytectl get launchplan
-----------------------

Gets the launch plan resources.

Synopsis
~~~~~~~~



Retrieve all launch plans within the project and domain:
::

 flytectl get launchplan -p flytesnacks -d development

.. note::
     The terms launchplan/launchplans are interchangeable in these commands.

 Retrieve a launch plan by name within the project and domain:

::

 flytectl get launchplan -p flytesnacks -d development core.basic.lp.go_greet


Retrieve the latest version of the task by name within the project and domain:

::

 flytectl get launchplan -p flytesnacks -d development  core.basic.lp.go_greet --latest

Retrieve a particular version of the launch plan by name within the project and domain:

::

 flytectl get launchplan -p flytesnacks -d development  core.basic.lp.go_greet --version v2

Retrieve all launch plans for a given workflow name:

::

 flytectl get launchplan -p flytesnacks -d development --workflow core.flyte_basics.lp.go_greet

Retrieve all the launch plans with filters:
::

  flytectl get launchplan -p flytesnacks -d development --filter.fieldSelector="name=core.basic.lp.go_greet"

Retrieve launch plans entity search across all versions with filters:
::

  flytectl get launchplan -p flytesnacks -d development k8s_spark.dataframe_passing.my_smart_schema --filter.fieldSelector="version=v1"


Retrieve all the launch plans with limit and sorting:
::

  flytectl get launchplan -p flytesnacks -d development --filter.sortBy=created_at --filter.limit=1 --filter.asc

Retrieve launch plans present in other pages by specifying the limit and page number:
::

  flytectl get -p flytesnacks -d development launchplan --filter.limit=10 --filter.page=2

Retrieve all launch plans within the project and domain in YAML format:

::

 flytectl get launchplan -p flytesnacks -d development -o yaml

Retrieve all launch plans the within the project and domain in JSON format:

::

 flytectl get launchplan -p flytesnacks -d development -o json

Retrieve a launch plan within the project and domain as per a version and generates the execution spec file; the file can be used to launch the execution using the 'create execution' command:

::

 flytectl get launchplan -d development -p flytectldemo core.advanced.run_merge_sort.merge_sort --execFile execution_spec.yaml

The generated file would look similar to this:

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

Check the :ref:`create execution section<flytectl_create_execution>` on how to launch one using the generated file.
Usage


::

  flytectl get launchplan [flags]

Options
~~~~~~~

::

      --execFile string               execution file name to be used for generating execution spec of a single launchplan.
      --filter.asc                    Specifies the sorting order. By default flytectl sort result in descending order
      --filter.fieldSelector string   Specifies the Field selector
      --filter.limit int32            Specifies the limit (default 100)
      --filter.page int32             Specifies the page number,  in case there are multiple pages of results (default 1)
      --filter.sortBy string          Specifies which field to sort results  (default "created_at")
  -h, --help                          help for launchplan
      --latest                         flag to indicate to fetch the latest version,  version flag will be ignored in this case
      --version string                version of the launchplan to be fetched.
      --workflow string               name of the workflow for which the launchplans need to be fetched.

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --admin.authorizationHeader string           Custom metadata header to pass JWT
      --admin.authorizationServerUrl string        This is the URL to your IdP's authorization server. It'll default to Endpoint
      --admin.caCertFilePath string                Use specified certificate file to verify the admin server peer.
      --admin.clientId string                      Client ID (default "flytepropeller")
      --admin.clientSecretLocation string          File containing the client secret (default "/etc/secrets/client_secret")
      --admin.command strings                      Command for external authentication token generation
      --admin.endpoint string                      For admin types,  specify where the uri of the service is located.
      --admin.insecure                             Use insecure connection.
      --admin.insecureSkipVerify                   InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name. Caution : shouldn't be use for production usecases'
      --admin.maxBackoffDelay string               Max delay for grpc backoff (default "8s")
      --admin.maxRetries int                       Max number of gRPC retries (default 4)
      --admin.perRetryTimeout string               gRPC per retry timeout (default "15s")
      --admin.pkceConfig.refreshTime string         (default "5m0s")
      --admin.pkceConfig.timeout string             (default "15s")
      --admin.scopes strings                       List of scopes to request
      --admin.tokenUrl string                      OPTIONAL: Your IdP's token endpoint. It'll be discovered from flyte admin's OAuth Metadata endpoint if not provided.
      --admin.useAuth                              Deprecated: Auth will be enabled/disabled based on admin's dynamically discovered information.
  -c, --config string                              config file (default is $HOME/.flyte/config.yaml)
  -d, --domain string                              Specifies the Flyte project's domain.
      --logger.formatter.type string               Sets logging format type. (default "json")
      --logger.level int                           Sets the minimum logging level. (default 4)
      --logger.mute                                Mutes all logs regardless of severity. Intended for benchmarks/tests only.
      --logger.show-source                         Includes source code location in logs.
  -o, --output string                              Specifies the output type - supported formats [TABLE JSON YAML DOT DOTURL]. NOTE: dot, doturl are only supported for Workflow (default "TABLE")
  -p, --project string                             Specifies the Flyte project.
      --storage.cache.max_size_mbs int             Maximum size of the cache where the Blob store data is cached in-memory. If not specified or set to 0,  cache is not used
      --storage.cache.target_gc_percent int        Sets the garbage collection target percentage.
      --storage.connection.access-key string       Access key to use. Only required when authtype is set to accesskey.
      --storage.connection.auth-type string        Auth Type to use [iam, accesskey]. (default "iam")
      --storage.connection.disable-ssl             Disables SSL connection. Should only be used for development.
      --storage.connection.endpoint string         URL for storage client to connect to.
      --storage.connection.region string           Region to connect to. (default "us-east-1")
      --storage.connection.secret-key string       Secret to use when accesskey is set.
      --storage.container string                   Initial container (in s3 a bucket) to create -if it doesn't exist-.'
      --storage.defaultHttpClient.timeout string   Sets time out on the http client. (default "0s")
      --storage.enable-multicontainer              If this is true,  then the container argument is overlooked and redundant. This config will automatically open new connections to new containers/buckets as they are encountered
      --storage.limits.maxDownloadMBs int          Maximum allowed download size (in MBs) per call. (default 2)
      --storage.stow.config stringToString         Configuration for stow backend. Refer to github/graymeta/stow (default [])
      --storage.stow.kind string                   Kind of Stow backend to use. Refer to github/graymeta/stow
      --storage.type string                        Sets the type of storage to configure [s3/minio/local/mem/stow]. (default "s3")

SEE ALSO
~~~~~~~~

* :doc:`flytectl_get` 	 - Fetches various Flyte resources such as tasks, workflows, launch plans, executions, and projects.

