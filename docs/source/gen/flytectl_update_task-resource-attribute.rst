.. _flytectl_update_task-resource-attribute:

flytectl update task-resource-attribute
---------------------------------------

Updates matchable resources of task attributes

Synopsis
~~~~~~~~



Updates task  resource attributes for given project and domain combination or additionally with workflow name.

Updating the task resource attribute is only available from a generated file. See the get section for generating this file.
Here the command updates takes the input for task resource attributes from the config file tra.yaml
eg:  content of tra.yaml

.. code-block:: yaml

	domain: development
	project: flytectldemo
	defaults:
		cpu: "1"
		memory: "150Mi"
	limits:
		cpu: "2"
		memory: "450Mi"

::

 flytectl update task-resource-attribute -attrFile tra.yaml

Updating task resource attribute for project and domain and workflow combination. This will take precedence over any other
resource attribute defined at project domain level.
Update the resource attributes for workflow core.control_flow.run_merge_sort.merge_sort in flytectldemo , development domain
.. code-block:: yaml

	domain: development
	project: flytectldemo
	workflow: core.control_flow.run_merge_sort.merge_sort
	defaults:
		cpu: "1"
		memory: "150Mi"
	limits:
		cpu: "2"
		memory: "450Mi"

::

 flytectl update task-resource-attribute -attrFile tra.yaml

Usage



::

  flytectl update task-resource-attribute [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for updating attribute for the resource type.
  -h, --help              help for task-resource-attribute

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

* :doc:`flytectl_update` 	 - Used for updating flyte resources eg: project.

