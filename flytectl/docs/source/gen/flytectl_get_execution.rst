.. _flytectl_get_execution:

flytectl get execution
----------------------

Gets execution resources.

Synopsis
~~~~~~~~



Retrieve all executions within the project and domain.
::

 flytectl get execution -p flytesnacks -d development

.. note::
    The terms execution/executions are interchangeable in these commands.

Retrieve executions by name within the project and domain.
::

 flytectl get execution -p flytesnacks -d development oeh94k9r2r

Retrieve all the executions with filters.
::

 flytectl get execution -p flytesnacks -d development --filter.fieldSelector="execution.phase in (FAILED;SUCCEEDED),execution.duration<200"


Retrieve executions as per the specified limit and sorting parameters.
::

 flytectl get execution -p flytesnacks -d development --filter.sortBy=created_at --filter.limit=1 --filter.asc

Retrieve executions present in other pages by specifying the limit and page number.

::

 flytectl get -p flytesnacks -d development execution --filter.limit=10 --filter.page=2

Retrieve executions within the project and domain in YAML format.

::

 flytectl get execution -p flytesnacks -d development -o yaml

Retrieve executions within the project and domain in JSON format.

::

 flytectl get execution -p flytesnacks -d development -o json


Get more details of the execution using the --details flag, which shows node and task executions. 
The default view is a tree view, and the TABLE view format is not supported on this view.

::

 flytectl get execution -p flytesnacks -d development oeh94k9r2r --details

Fetch execution details in YAML format. In this view, only node details are available. For task, pass the --nodeID flag.
::

 flytectl get execution -p flytesnacks -d development oeh94k9r2r --details -o yaml

Fetch task executions on a specific node using the --nodeID flag. Use the nodeID attribute given by the node details view.

::

 flytectl get execution -p flytesnacks -d development oeh94k9r2r --nodeID n0

Task execution view is available in YAML/JSON format too. The following example showcases YAML, where the output contains input and output data of each node.

::

 flytectl get execution -p flytesnacks -d development oeh94k9r2r --nodID n0 -o yaml

Usage


::

  flytectl get execution [flags]

Options
~~~~~~~

::

      --details                       gets node execution details. Only applicable for single execution name i.e get execution name --details
      --filter.asc                    Specifies the sorting order. By default flytectl sort result in descending order
      --filter.fieldSelector string   Specifies the Field selector
      --filter.limit int32            Specifies the limit (default 100)
      --filter.page int32             Specifies the page number,  in case there are multiple pages of results (default 1)
      --filter.sortBy string          Specifies which field to sort results  (default "created_at")
  -h, --help                          help for execution
      --nodeID string                 get task executions for given node name.

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --admin.audience string                        Audience to use when initiating OAuth2 authorization requests.
      --admin.authType string                        Type of OAuth2 flow used for communicating with admin.ClientSecret, Pkce, ExternalCommand are valid values (default "ClientSecret")
      --admin.authorizationHeader string             Custom metadata header to pass JWT
      --admin.authorizationServerUrl string          This is the URL to your IdP's authorization server. It'll default to Endpoint
      --admin.caCertFilePath string                  Use specified certificate file to verify the admin server peer.
      --admin.clientId string                        Client ID (default "flytepropeller")
      --admin.clientSecretEnvVar string              Environment variable containing the client secret
      --admin.clientSecretLocation string            File containing the client secret (default "/etc/secrets/client_secret")
      --admin.command strings                        Command for external authentication token generation
      --admin.defaultServiceConfig string            
      --admin.deviceFlowConfig.pollInterval string   amount of time the device flow would poll the token endpoint if auth server doesn't return a polling interval. Okta and google IDP do return an interval' (default "5s")
      --admin.deviceFlowConfig.refreshTime string    grace period from the token expiry after which it would refresh the token. (default "5m0s")
      --admin.deviceFlowConfig.timeout string        amount of time the device flow should complete or else it will be cancelled. (default "10m0s")
      --admin.endpoint string                        For admin types,  specify where the uri of the service is located.
      --admin.httpProxyURL string                    OPTIONAL: HTTP Proxy to be used for OAuth requests.
      --admin.insecure                               Use insecure connection.
      --admin.insecureSkipVerify                     InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name. Caution : shouldn't be use for production usecases'
      --admin.maxBackoffDelay string                 Max delay for grpc backoff (default "8s")
      --admin.maxRetries int                         Max number of gRPC retries (default 4)
      --admin.perRetryTimeout string                 gRPC per retry timeout (default "15s")
      --admin.pkceConfig.refreshTime string          grace period from the token expiry after which it would refresh the token. (default "5m0s")
      --admin.pkceConfig.timeout string              Amount of time the browser session would be active for authentication from client app. (default "2m0s")
      --admin.scopes strings                         List of scopes to request
      --admin.tokenRefreshWindow string              Max duration between token refresh attempt and token expiry. (default "0s")
      --admin.tokenUrl string                        OPTIONAL: Your IdP's token endpoint. It'll be discovered from flyte admin's OAuth Metadata endpoint if not provided.
      --admin.useAudienceFromAdmin                   Use Audience configured from admins public endpoint config.
      --admin.useAuth                                Deprecated: Auth will be enabled/disabled based on admin's dynamically discovered information.
  -c, --config string                                config file (default is $HOME/.flyte/config.yaml)
      --console.endpoint string                      Endpoint of console,  if different than flyte admin
  -d, --domain string                                Specifies the Flyte project's domain.
      --files.archive                                Pass in archive file either an http link or local path.
      --files.assumableIamRole string                Custom assumable iam auth role to register launch plans with.
      --files.continueOnError                        Continue on error when registering files.
      --files.destinationDirectory string            Location of source code in container.
      --files.dryRun                                 Execute command without making any modifications.
      --files.enableSchedule                         Enable the schedule if the files contain schedulable launchplan.
      --files.force                                  Force use of version number on entities registered with flyte.
      --files.k8ServiceAccount string                Deprecated. Please use --K8sServiceAccount
      --files.k8sServiceAccount string               Custom kubernetes service account auth role to register launch plans with.
      --files.outputLocationPrefix string            Custom output location prefix for offloaded types (files/schemas).
      --files.sourceUploadPath string                Deprecated: Update flyte admin to avoid having to configure storage access from flytectl.
      --files.version string                         Version of the entity to be registered with flyte which are un-versioned after serialization.
      --logger.formatter.type string                 Sets logging format type. (default "json")
      --logger.level int                             Sets the minimum logging level. (default 3)
      --logger.mute                                  Mutes all logs regardless of severity. Intended for benchmarks/tests only.
      --logger.show-source                           Includes source code location in logs.
  -o, --output string                                Specifies the output type - supported formats [TABLE JSON YAML DOT DOTURL]. NOTE: dot, doturl are only supported for Workflow (default "TABLE")
  -p, --project string                               Specifies the Flyte project.
      --storage.cache.max_size_mbs int               Maximum size of the cache where the Blob store data is cached in-memory. If not specified or set to 0,  cache is not used
      --storage.cache.target_gc_percent int          Sets the garbage collection target percentage.
      --storage.connection.access-key string         Access key to use. Only required when authtype is set to accesskey.
      --storage.connection.auth-type string          Auth Type to use [iam, accesskey]. (default "iam")
      --storage.connection.disable-ssl               Disables SSL connection. Should only be used for development.
      --storage.connection.endpoint string           URL for storage client to connect to.
      --storage.connection.region string             Region to connect to. (default "us-east-1")
      --storage.connection.secret-key string         Secret to use when accesskey is set.
      --storage.container string                     Initial container (in s3 a bucket) to create -if it doesn't exist-.'
      --storage.defaultHttpClient.timeout string     Sets time out on the http client. (default "0s")
      --storage.enable-multicontainer                If this is true,  then the container argument is overlooked and redundant. This config will automatically open new connections to new containers/buckets as they are encountered
      --storage.limits.maxDownloadMBs int            Maximum allowed download size (in MBs) per call. (default 2)
      --storage.stow.config stringToString           Configuration for stow backend. Refer to github/flyteorg/stow (default [])
      --storage.stow.kind string                     Kind of Stow backend to use. Refer to github/flyteorg/stow
      --storage.type string                          Sets the type of storage to configure [s3/minio/local/mem/stow]. (default "s3")

SEE ALSO
~~~~~~~~

* :doc:`flytectl_get` 	 - Fetches various Flyte resources such as tasks, workflows, launch plans, executions, and projects.

