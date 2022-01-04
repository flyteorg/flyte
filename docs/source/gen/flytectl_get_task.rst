.. _flytectl_get_task:

flytectl get task
-----------------

Get task resources

Synopsis
~~~~~~~~



Retrieve all the tasks within project and domain(task,tasks can be used interchangeably in these commands):
::

 flytectl get task -p flytesnacks -d development

Retrieve task by name within project and domain:

::

 flytectl task -p flytesnacks -d development core.basic.lp.greet

Retrieve latest version of task by name within project and domain:

::

 flytectl get task -p flytesnacks -d development  core.basic.lp.greet --latest

Retrieve particular version of task by name within project and domain:

::

 flytectl get task -p flytesnacks -d development  core.basic.lp.greet --version v2

Retrieve all the tasks with filters:
::

  flytectl get task -p flytesnacks -d development --filter.fieldSelector="task.name=k8s_spark.pyspark_pi.print_every_time,task.version=v1"

Retrieve a specific task with filters:
::

  flytectl get task -p flytesnacks -d development k8s_spark.pyspark_pi.print_every_time --filter.fieldSelector="task.version=v1,created_at>=2021-05-24T21:43:12.325335Z"

Retrieve all the tasks with limit and sorting:
::

  flytectl get -p flytesnacks -d development task  --filter.sortBy=created_at --filter.limit=1 --filter.asc

Retrieve tasks present in other pages by specifying the limit and page number:
::

  flytectl get -p flytesnacks -d development task --filter.limit=10 --filter.page=2

Retrieve all the tasks within project and domain in yaml format:
::

 flytectl get task -p flytesnacks -d development -o yaml

Retrieve all the tasks within project and domain in json format:

::

 flytectl get task -p flytesnacks -d development -o json

Retrieve tasks within project and domain for a version and generate the execution spec file for it to be used for launching the execution using create execution:

::

 flytectl get tasks -d development -p flytesnacks core.advanced.run_merge_sort.merge --execFile execution_spec.yaml --version v2

The generated file would look similar to this:

.. code-block:: yaml

	 iamRoleARN: ""
	 inputs:
	   sorted_list1:
	   - 0
	   sorted_list2:
	   - 0
	 kubeServiceAcct: ""
	 targetDomain: ""
	 targetProject: ""
	 task: core.advanced.run_merge_sort.merge
	 version: v2

Check the create execution section on how to launch one using the generated file.

Usage


::

  flytectl get task [flags]

Options
~~~~~~~

::

      --execFile string               execution file name to be used for generating execution spec of a single task.
      --filter.asc                    Specifies the sorting order. By default flytectl sort result in descending order
      --filter.fieldSelector string   Specifies the Field selector
      --filter.limit int32            Specifies the limit (default 100)
      --filter.page int32             Specifies the page number,  in case there are multiple pages of results (default 1)
      --filter.sortBy string          Specifies which field to sort results  (default "created_at")
  -h, --help                          help for task
      --latest                         flag to indicate to fetch the latest version,  version flag will be ignored in this case
      --version string                version of the task to be fetched.

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

* :doc:`flytectl_get` 	 - Fetch various Flyte resources including tasks/workflows/launchplans/executions/project.

