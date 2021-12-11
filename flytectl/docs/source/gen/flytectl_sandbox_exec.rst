.. _flytectl_sandbox_exec:

flytectl sandbox exec
---------------------

Execute non-interactive command inside the sandbox container

Synopsis
~~~~~~~~



Runs non-interactive command inside the Sandbox container and immediately returns the output. By default, flytectl exec is present in /root directory inside the Sandbox container.

::
 flytectl sandbox exec -- ls -al 

Usage

::

  flytectl sandbox exec [flags]

Options
~~~~~~~

::

  -h, --help   help for exec

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --admin.authorizationHeader string           Custom metadata header to pass JWT
      --admin.authorizationServerUrl string        This is the URL to your IdP's authorization server. It'll default to Endpoint
      --admin.clientId string                      Client ID (default "flytepropeller")
      --admin.clientSecretLocation string          File containing the client secret (default "/etc/secrets/client_secret")
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

* :doc:`flytectl_sandbox` 	 - Used for sandbox interactions like start/teardown/status/exec.

