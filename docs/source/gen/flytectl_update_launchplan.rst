.. _flytectl_update_launchplan:

flytectl update launchplan
--------------------------

Updates launch plan metadata

Synopsis
~~~~~~~~



Following command updates the description on the launchplan.
::

 flytectl update launchplan -p flytectldemo -d development  core.advanced.run_merge_sort.merge_sort --description "Mergesort example"

Archiving launchplan named entity is not supported and would throw an error.
::

 flytectl update launchplan -p flytectldemo -d development  core.advanced.run_merge_sort.merge_sort --archive

Activating launchplan named entity would be a noop.
::

 flytectl update launchplan -p flytectldemo -d development  core.advanced.run_merge_sort.merge_sort --activate

Usage


::

  flytectl update launchplan [flags]

Options
~~~~~~~

::

      --activate             Activates the named entity specified as argument.
      --archive              Archives the named entity specified as argument.
      --description string   description of the namedentity.
  -h, --help                 help for launchplan

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --admin.authorizationHeader string           Custom metadata header to pass JWT
      --admin.authorizationServerUrl string        This is the URL to your IDP's authorization server'
      --admin.clientId string                      Client ID
      --admin.clientSecretLocation string          File containing the client secret
      --admin.endpoint string                      For admin types,  specify where the uri of the service is located.
      --admin.insecure                             Use insecure connection.
      --admin.maxBackoffDelay string               Max delay for grpc backoff (default "8s")
      --admin.maxRetries int                       Max number of gRPC retries (default 4)
      --admin.perRetryTimeout string               gRPC per retry timeout (default "15s")
      --admin.scopes strings                       List of scopes to request
      --admin.tokenUrl string                      Your IDPs token endpoint
      --admin.useAuth                              Whether or not to try to authenticate with options below
      --adminutils.batchSize int                   Maximum number of records to retrieve per call. (default 100)
      --adminutils.maxRecords int                  Maximum number of records to retrieve. (default 500)
      --config string                              config file (default is $HOME/config.yaml)
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

