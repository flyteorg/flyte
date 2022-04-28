.. _flytectl_delete_workflow-execution-config:

flytectl delete workflow-execution-config
-----------------------------------------

Deletes matchable resources of workflow execution config.

Synopsis
~~~~~~~~



Delete workflow execution config for the given project and domain combination or additionally the workflow name.

For project flytectldemo and development domain, run:
::

 flytectl delete workflow-execution-config -p flytectldemo -d development

To delete workflow execution config using the config file which was used to create it, run:

::

 flytectl delete workflow-execution-config --attrFile wec.yaml

For example, here's the config file wec.yaml:

.. code-block:: yaml

    domain: development
    project: flytectldemo
    max_parallelism: 5
	security_context:
  		run_as:
    		k8s_service_account: demo

Max_parallelism is optional in the file as it is unread during the delete command but can be retained since the same file can be used for get, update and delete commands.

To delete workflow execution config for the workflow 'core.control_flow.run_merge_sort.merge_sort', run:

::

 flytectl delete workflow-execution-config -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

Usage


::

  flytectl delete workflow-execution-config [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for delete attribute for the resource type.
      --dryRun            execute command without making any modifications.
  -h, --help              help for workflow-execution-config

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

* :doc:`flytectl_delete` 	 - Terminates/deletes various Flyte resources such as tasks, workflows, launch plans, executions, and projects.

