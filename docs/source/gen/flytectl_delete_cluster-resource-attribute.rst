.. _flytectl_delete_cluster-resource-attribute:

flytectl delete cluster-resource-attribute
------------------------------------------

Deletes matchable resources of cluster attributes

Synopsis
~~~~~~~~



Deletes cluster resource attributes for given project and domain combination or additionally with workflow name.

Deletes cluster resource attribute for project and domain
Here the command delete cluster resource attributes for  project flytectldemo and development domain.
::

 flytectl delete cluster-resource-attribute -p flytectldemo -d development 


Deletes cluster resource attribute using config file which was used for creating it.
Here the command deletes cluster resource attributes from the config file cra.yaml
Attributes are optional in the file as they are unread during the delete command but can be kept as the same file can be used for get, update or delete 
eg:  content of cra.yaml which will use the project domain and workflow name for deleting the resource

::

 flytectl delete cluster-resource-attribute --attrFile cra.yaml


.. code-block:: yaml
	
    domain: development
    project: flytectldemo
    attributes:
      foo: "bar"
      buzz: "lightyear"

Deletes cluster resource attribute for a workflow
Here the command deletes cluster resource attributes for a workflow

::

 flytectl delete cluster-resource-attribute -p flytectldemo -d development core.control_flow.run_merge_sort.merge_sort

Usage


::

  flytectl delete cluster-resource-attribute [flags]

Options
~~~~~~~

::

      --attrFile string   attribute file name to be used for delete attribute for the resource type.
  -h, --help              help for cluster-resource-attribute

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

* :doc:`flytectl_delete` 	 - Used for terminating/deleting various flyte resources including tasks/workflows/launchplans/executions/project.

