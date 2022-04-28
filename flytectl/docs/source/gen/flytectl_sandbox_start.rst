.. _flytectl_sandbox_start:

flytectl sandbox start
----------------------

Starts the Flyte sandbox cluster.

Synopsis
~~~~~~~~



Flyte sandbox is a fully standalone minimal environment for running Flyte.
It provides a simplified way of running Flyte sandbox as a single Docker container locally.

Starts the sandbox cluster without any source code:
::

 flytectl sandbox start

Mounts your source code repository inside the sandbox:
::

 flytectl sandbox start --source=$HOME/flyteorg/flytesnacks

Runs a specific version of Flyte. Flytectl sandbox only supports Flyte version available in the Github release, https://github.com/flyteorg/flyte/tags.
::

 flytectl sandbox start  --version=v0.14.0

.. note::
	  Flytectl Sandbox is only supported for Flyte versions > v0.10.0.

Runs the latest pre release of  Flyte.
::

 flytectl sandbox start  --pre

Note: The pre release flag will be ignored if the user passes the version flag. In that case, Flytectl will use a specific version.

Specify a Flyte Sandbox compliant image with the registry. This is useful in case you want to use an image from your registry.
::

  flytectl sandbox start --image docker.io/my-override:latest

Note: If image flag is passed then Flytectl will ignore version and pre flags.

Specify a Flyte Sandbox image pull policy. Possible pull policy values are Always, IfNotPresent, or Never:
::

 flytectl sandbox start  --image docker.io/my-override:latest --imagePullPolicy Always

Start sandbox cluster passing environment variables. This can be used to pass docker specific env variables or flyte specific env variables.
eg : for passing timeout value in secs for the sandbox container use the following.
::

 flytectl sandbox start --env FLYTE_TIMEOUT=700


The DURATION can be a positive integer or a floating-point number, followed by an optional unit suffix::
s - seconds (default)
m - minutes
h - hours
d - days
When no unit is used, it defaults to seconds. If the duration is set to zero, the associated timeout is disabled.


eg : for passing multiple environment variables
::

 flytectl sandbox start --env USER=foo --env PASSWORD=bar


Usage


::

  flytectl sandbox start [flags]

Options
~~~~~~~

::

      --env strings                            Optional. Provide Env variable in key=value format which can be passed to sandbox container.
  -h, --help                                   help for start
      --image string                           Optional. Provide a fully qualified path to a Flyte compliant docker image.
      --imagePullOptions.platform string       Forces a specific platform's image to be pulled.'
      --imagePullOptions.registryAuth string   The base64 encoded credentials for the registry.
      --imagePullPolicy ImagePullPolicy        Optional. Defines the image pull behavior [Always/IfNotPresent/Never] (default Always)
      --pre                                    Optional. Pre release Version of flyte will be used for sandbox.
      --source string                          Path of your source code
      --version string                         Version of flyte. Only supports flyte releases greater than v0.10.0

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --admin.authType string                      Type of OAuth2 flow used for communicating with admin.ClientSecret, Pkce, ExternalCommand are valid values (default "ClientSecret")
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
      --admin.tokenRefreshWindow string            Max duration between token refresh attempt and token expiry. (default "0s")
      --admin.tokenUrl string                      OPTIONAL: Your IdP's token endpoint. It'll be discovered from flyte admin's OAuth Metadata endpoint if not provided.
      --admin.useAuth                              Deprecated: Auth will be enabled/disabled based on admin's dynamically discovered information.
  -c, --config string                              config file (default is $HOME/.flyte/config.yaml)
  -d, --domain string                              Specifies the Flyte project's domain.
      --files.archive                              Pass in archive file either an http link or local path.
      --files.assumableIamRole string              Custom assumable iam auth role to register launch plans with.
      --files.continueOnError                      Continue on error when registering files.
      --files.destinationDirectory string          Location of source code in container.
      --files.dryRun                               Execute command without making any modifications.
      --files.force                                Force use of version number on entities registered with flyte.
      --files.k8ServiceAccount string              Deprecated. Please use --K8sServiceAccount
      --files.k8sServiceAccount string             Custom kubernetes service account auth role to register launch plans with.
      --files.outputLocationPrefix string          Custom output location prefix for offloaded types (files/schemas).
      --files.sourceUploadPath string              Deprecated: Update flyte admin to avoid having to configure storage access from flytectl.
      --files.version string                       Version of the entity to be registered with flyte which are un-versioned after serialization.
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

* :doc:`flytectl_sandbox` 	 - Helps with sandbox interactions like start, teardown, status, and exec.

