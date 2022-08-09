.. _flytectl_update_task-resource-attribute:

flytectl update task-resource-attribute
---------------------------------------

Update matchable resources of task attributes

Synopsis
~~~~~~~~



Updates the task resource attributes for the given project and domain combination or additionally with workflow name.

Updating the task resource attribute is only available from a generated file. See the get section for generating this file.
This will completely overwrite any existing custom project, domain, and workflow combination attributes.
It is preferable to do get and generate an attribute file if there is an existing attribute already set and then update it to have new values.
Refer to get task-resource-attribute section on how to generate this file.
It takes input for task resource attributes from the config file tra.yaml,
Example: content of tra.yaml:

.. code-block:: yaml

    domain: development
    project: flytesnacks
    defaults:
      cpu: "1"
      memory: "150Mi"
    limits:
      cpu: "2"
      memory: "450Mi"

::

 flytectl update task-resource-attribute --attrFile tra.yaml

Update task resource attribute for project, domain, and workflow combination. This will take precedence over any other
resource attribute defined at project domain level.
For workflow 'core.control_flow.merge_sort.merge_sort' in flytesnacks project, development domain, it is:

.. code-block:: yaml

    domain: development
    project: flytesnacks
    workflow: core.control_flow.merge_sort.merge_sort
    defaults:
      cpu: "1"
      memory: "150Mi"
    limits:
      cpu: "2"
      memory: "450Mi"

::

 flytectl update task-resource-attribute --attrFile tra.yaml

Usage



::

  flytectl update task-resource-attribute [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for updating attribute for the resource type.
      --dryRun            execute command without making any modifications.
  -h, --help              help for task-resource-attribute

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
      --admin.defaultServiceConfig string          
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
      --files.enableSchedule                       Enable the schedule if the files contain schedulable launchplan.
      --files.force                                Force use of version number on entities registered with flyte.
      --files.k8ServiceAccount string              Deprecated. Please use --K8sServiceAccount
      --files.k8sServiceAccount string             Custom kubernetes service account auth role to register launch plans with.
      --files.outputLocationPrefix string          Custom output location prefix for offloaded types (files/schemas).
      --files.sourceUploadPath string              Deprecated: Update flyte admin to avoid having to configure storage access from flytectl.
      --files.version string                       Version of the entity to be registered with flyte which are un-versioned after serialization.
      --logger.formatter.type string               Sets logging format type. (default "json")
      --logger.level int                           Sets the minimum logging level. (default 3)
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
      --storage.stow.config stringToString         Configuration for stow backend. Refer to github/flyteorg/stow (default [])
      --storage.stow.kind string                   Kind of Stow backend to use. Refer to github/flyteorg/stow
      --storage.type string                        Sets the type of storage to configure [s3/minio/local/mem/stow]. (default "s3")

SEE ALSO
~~~~~~~~

* :doc:`flytectl_update` 	 - Update Flyte resources e.g., project.

