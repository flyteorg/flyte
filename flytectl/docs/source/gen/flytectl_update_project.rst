.. _flytectl_update_project:

flytectl update project
-----------------------

Update the characteristics of a project

Synopsis
~~~~~~~~



Allows you to update the characteristics of a project, including its name, labels and description.
Also allows you to archive or activate (unarchive) a project.

To archive a project, specify its ID with the *p* flag and add the *archive* flag:

::

    flytectl update project -p my-project-id --archive

To activate (unarchive) an archived project, specify its ID with the *p* flag and add the *activate* flag:

::

	flytectl update project -p my-project-id --activate

To update the characteristics of a project using flags, specify the project ID with the *p* flag and the flags corresponding to the characteristics you want to update:

::

    flytectl update project -p my-project-id --description "A wonderful project"  --labels app=my-app

To update the characteristics of a project using a *yaml* file, define the file with the project ID desired updates:

.. code-block:: yaml

    id: "my-project-id"
    name: "my-project-name"
    labels:
       values:
         app: my-app
    description: "A wonderful project"


(Note: The name parameter must not contain whitespace)

Then, pass it in using the *file* flag:

::

    flytectl update project --file project.yaml

To archive or activate (unarchive) a project using a *yaml* file:

* Add a state field, with a value of *0* for activated (unarchived) or *1* for archived, at the top level of the *yaml* file.

* Add the *archive* flag to the command.

For example, to archive a project:

.. code-block:: yaml

    # update.yaml
    id: "my-project-id"
    state: 1

::

    $ flytectl update project --file update.yaml --archive

And to activate (unarchive) the same project:

.. code-block:: yaml

    # update.yaml
    id: "my-project-id"
    state: 0

::

    $ flytectl update project --file update.yaml --archive

Note that when using a *yaml* file, the *activate* flag is not used.
Instead, the *archive* flag is used for *both* archiving and activating (unarchiving) with the difference being in the *state* field of the *yaml* file.
Furthermore, the *state* field only takes effect if the *archive* flag is present in the command.

Usage


::

  flytectl update project [flags]

Options
~~~~~~~

::

      --activate                Activates the project specified as argument. Only used in update
      --activateProject         (Deprecated) Activates the project specified as argument. Only used in update
      --archive                 Archives the project specified as argument. Only used in update
      --archiveProject          (Deprecated) Archives the project specified as argument. Only used in update
      --description string      description for the project specified as argument.
      --dryRun                  execute command without making any modifications.
      --file string             file for the project definition.
      --force                   Skips asking for an acknowledgement during an update operation. Only used in update
  -h, --help                    help for project
      --id string               id for the project specified as argument.
      --labels stringToString   labels for the project specified as argument. (default [])
      --name string             name for the project specified as argument.

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

* :doc:`flytectl_update` 	 - Update Flyte resources e.g., project.

