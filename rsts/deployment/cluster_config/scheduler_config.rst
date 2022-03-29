.. _flytescheduler-config-specification:

#########################################
Flyte Scheduler Configuration
#########################################

- `admin <#section-admin>`_

- `cluster_resources <#section-cluster_resources>`_

- `clusters <#section-clusters>`_

- `database <#section-database>`_

- `domains <#section-domains>`_

- `externalevents <#section-externalevents>`_

- `flyteadmin <#section-flyteadmin>`_

- `logger <#section-logger>`_

- `namespace_mapping <#section-namespace_mapping>`_

- `notifications <#section-notifications>`_

- `qualityofservice <#section-qualityofservice>`_

- `queues <#section-queues>`_

- `registration <#section-registration>`_

- `remotedata <#section-remotedata>`_

- `scheduler <#section-scheduler>`_

- `storage <#section-storage>`_

- `task_resources <#section-task_resources>`_

- `task_type_whitelist <#section-task_type_whitelist>`_

Section: admin
================================================================================

endpoint (`config.URL`_)
--------------------------------------------------------------------------------

For admin types, specify where the uri of the service is located.

**Default Value**: 

.. code-block:: yaml

  ""
  

insecure (bool)
--------------------------------------------------------------------------------

Use insecure connection.

**Default Value**: 

.. code-block:: yaml

  "false"
  

insecureSkipVerify (bool)
--------------------------------------------------------------------------------

InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name. Caution : shouldn't be use for production usecases'

**Default Value**: 

.. code-block:: yaml

  "false"
  

caCertFilePath (string)
--------------------------------------------------------------------------------

Use specified certificate file to verify the admin server peer.

**Default Value**: 

.. code-block:: yaml

  ""
  

maxBackoffDelay (`config.Duration`_)
--------------------------------------------------------------------------------

Max delay for grpc backoff

**Default Value**: 

.. code-block:: yaml

  8s
  

perRetryTimeout (`config.Duration`_)
--------------------------------------------------------------------------------

gRPC per retry timeout

**Default Value**: 

.. code-block:: yaml

  15s
  

maxRetries (int)
--------------------------------------------------------------------------------

Max number of gRPC retries

**Default Value**: 

.. code-block:: yaml

  "4"
  

authType (uint8)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ClientSecret
  

useAuth (bool)
--------------------------------------------------------------------------------

Deprecated: Auth will be enabled/disabled based on admin's dynamically discovered information.

**Default Value**: 

.. code-block:: yaml

  "false"
  

clientId (string)
--------------------------------------------------------------------------------

Client ID

**Default Value**: 

.. code-block:: yaml

  flytepropeller
  

clientSecretLocation (string)
--------------------------------------------------------------------------------

File containing the client secret

**Default Value**: 

.. code-block:: yaml

  /etc/secrets/client_secret
  

scopes ([]string)
--------------------------------------------------------------------------------

List of scopes to request

**Default Value**: 

.. code-block:: yaml

  []
  

authorizationServerUrl (string)
--------------------------------------------------------------------------------

This is the URL to your IdP's authorization server. It'll default to Endpoint

**Default Value**: 

.. code-block:: yaml

  ""
  

tokenUrl (string)
--------------------------------------------------------------------------------

OPTIONAL: Your IdP's token endpoint. It'll be discovered from flyte admin's OAuth Metadata endpoint if not provided.

**Default Value**: 

.. code-block:: yaml

  ""
  

authorizationHeader (string)
--------------------------------------------------------------------------------

Custom metadata header to pass JWT

**Default Value**: 

.. code-block:: yaml

  ""
  

pkceConfig (`pkce.Config`_)
--------------------------------------------------------------------------------

Config for Pkce authentication flow.

**Default Value**: 

.. code-block:: yaml

  refreshTime: 5m0s
  timeout: 15s
  

command ([]string)
--------------------------------------------------------------------------------

Command for external authentication token generation

**Default Value**: 

.. code-block:: yaml

  []
  

config.Duration
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Duration (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  8s
  

config.URL
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

URL (`url.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ForceQuery: false
  Fragment: ""
  Host: ""
  Opaque: ""
  Path: ""
  RawFragment: ""
  RawPath: ""
  RawQuery: ""
  Scheme: ""
  User: null
  

url.URL
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Scheme (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Opaque (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

User (url.Userinfo)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

Host (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Path (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

RawPath (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

ForceQuery (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "false"
  

RawQuery (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Fragment (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

RawFragment (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

pkce.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

timeout (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  15s
  

refreshTime (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  5m0s
  

Section: cluster_resources
================================================================================

templatePath (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

templateData (map[string]interfaces.DataSource)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  {}
  

refreshInterval (`config.Duration`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  1m0s
  

customData (map[string]map[string]interfaces.DataSource)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  {}
  

standaloneDeployment (bool)
--------------------------------------------------------------------------------

Whether the cluster resource sync is running in a standalone deployment and should call flyteadmin service endpoints

**Default Value**: 

.. code-block:: yaml

  "false"
  

Section: clusters
================================================================================

clusterConfigs ([]interfaces.ClusterConfig)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  null
  

labelClusterMap (map[string][]interfaces.ClusterEntity)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  null
  

Section: database
================================================================================

host (string)
--------------------------------------------------------------------------------

deprecated

**Default Value**: 

.. code-block:: yaml

  postgres
  

port (int)
--------------------------------------------------------------------------------

deprecated

**Default Value**: 

.. code-block:: yaml

  "5432"
  

dbname (string)
--------------------------------------------------------------------------------

deprecated

**Default Value**: 

.. code-block:: yaml

  postgres
  

username (string)
--------------------------------------------------------------------------------

deprecated

**Default Value**: 

.. code-block:: yaml

  postgres
  

password (string)
--------------------------------------------------------------------------------

deprecated

**Default Value**: 

.. code-block:: yaml

  ""
  

passwordPath (string)
--------------------------------------------------------------------------------

deprecated

**Default Value**: 

.. code-block:: yaml

  ""
  

options (string)
--------------------------------------------------------------------------------

deprecated

**Default Value**: 

.. code-block:: yaml

  sslmode=disable
  

debug (bool)
--------------------------------------------------------------------------------

deprecated

**Default Value**: 

.. code-block:: yaml

  "false"
  

enableForeignKeyConstraintWhenMigrating (bool)
--------------------------------------------------------------------------------

Whether to enable gorm foreign keys when migrating the db

**Default Value**: 

.. code-block:: yaml

  "false"
  

maxIdleConnections (int)
--------------------------------------------------------------------------------

maxIdleConnections sets the maximum number of connections in the idle connection pool.

**Default Value**: 

.. code-block:: yaml

  "10"
  

maxOpenConnections (int)
--------------------------------------------------------------------------------

maxOpenConnections sets the maximum number of open connections to the database.

**Default Value**: 

.. code-block:: yaml

  "1000"
  

connMaxLifeTime (`config.Duration`_)
--------------------------------------------------------------------------------

sets the maximum amount of time a connection may be reused

**Default Value**: 

.. code-block:: yaml

  1h0m0s
  

postgres (`interfaces.PostgresConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  dbname: ""
  debug: false
  host: ""
  options: ""
  password: ""
  passwordPath: ""
  port: 0
  username: ""
  

interfaces.PostgresConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

host (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

The host name of the database server

**Default Value**: 

.. code-block:: yaml

  ""
  

port (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

The port name of the database server

**Default Value**: 

.. code-block:: yaml

  "0"
  

dbname (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

The database name

**Default Value**: 

.. code-block:: yaml

  ""
  

username (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

The database user who is connecting to the server.

**Default Value**: 

.. code-block:: yaml

  ""
  

password (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

The database password.

**Default Value**: 

.. code-block:: yaml

  ""
  

passwordPath (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Points to the file containing the database password.

**Default Value**: 

.. code-block:: yaml

  ""
  

options (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

See http://gorm.io/docs/connecting_to_the_database.html for available options passed, in addition to the above.

**Default Value**: 

.. code-block:: yaml

  ""
  

debug (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Whether or not to start the database connection with debug mode enabled.

**Default Value**: 

.. code-block:: yaml

  "false"
  

Section: domains
================================================================================

id (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  development
  

name (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  development
  

Section: externalevents
================================================================================

enable (bool)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "false"
  

type (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  local
  

aws (`interfaces.AWSConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  region: ""
  

gcp (`interfaces.GCPConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  projectId: ""
  

eventsPublisher (`interfaces.EventsPublisherConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  eventTypes: null
  topicName: ""
  

reconnectAttempts (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

reconnectDelaySeconds (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

interfaces.AWSConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

region (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

interfaces.EventsPublisherConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

topicName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

eventTypes ([]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

interfaces.GCPConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

projectId (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Section: flyteadmin
================================================================================

roleNameKey (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

metricsScope (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  'flyte:'
  

profilerPort (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "10254"
  

metadataStoragePrefix ([]string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  - metadata
  - admin
  

eventVersion (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "2"
  

asyncEventsBufferSize (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "100"
  

maxParallelism (int32)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "25"
  

Section: logger
================================================================================

show-source (bool)
--------------------------------------------------------------------------------

Includes source code location in logs.

**Default Value**: 

.. code-block:: yaml

  "false"
  

mute (bool)
--------------------------------------------------------------------------------

Mutes all logs regardless of severity. Intended for benchmarks/tests only.

**Default Value**: 

.. code-block:: yaml

  "false"
  

level (int)
--------------------------------------------------------------------------------

Sets the minimum logging level.

**Default Value**: 

.. code-block:: yaml

  "4"
  

formatter (`logger.FormatterConfig`_)
--------------------------------------------------------------------------------

Sets logging format.

**Default Value**: 

.. code-block:: yaml

  type: json
  

logger.FormatterConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

type (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Sets logging format type.

**Default Value**: 

.. code-block:: yaml

  json
  

Section: namespace_mapping
================================================================================

mapping (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

template (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  '{{ project }}-{{ domain }}'
  

templateData (map[string]interfaces.DataSource)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  null
  

Section: notifications
================================================================================

type (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  local
  

region (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

aws (`interfaces.AWSConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  region: ""
  

gcp (`interfaces.GCPConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  projectId: ""
  

publisher (`interfaces.NotificationsPublisherConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  topicName: ""
  

processor (`interfaces.NotificationsProcessorConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  accountId: ""
  queueName: ""
  

emailer (`interfaces.NotificationsEmailerConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  body: ""
  emailServerConfig:
    apiKeyEnvVar: ""
    apiKeyFilePath: ""
    serviceName: ""
  sender: ""
  subject: ""
  

reconnectAttempts (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

reconnectDelaySeconds (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

interfaces.NotificationsEmailerConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

emailServerConfig (`interfaces.EmailServerConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  apiKeyEnvVar: ""
  apiKeyFilePath: ""
  serviceName: ""
  

subject (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

sender (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

body (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

interfaces.EmailServerConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

serviceName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

apiKeyEnvVar (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

apiKeyFilePath (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

interfaces.NotificationsProcessorConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

queueName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

accountId (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

interfaces.NotificationsPublisherConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

topicName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Section: qualityofservice
================================================================================

tierExecutionValues (map[string]interfaces.QualityOfServiceSpec)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  {}
  

defaultTiers (map[string]string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  {}
  

Section: queues
================================================================================

executionQueues (interfaces.ExecutionQueues)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  []
  

workflowConfigs (interfaces.WorkflowConfigs)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  []
  

Section: registration
================================================================================

maxWorkflowNodes (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "100"
  

maxLabelEntries (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

maxAnnotationEntries (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

workflowSizeLimit (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

Section: remotedata
================================================================================

scheme (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  none
  

region (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

signedUrls (`interfaces.SignedURL`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  durationMinutes: 0
  enabled: false
  signingPrincipal: ""
  

maxSizeInBytes (int64)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "2097152"
  

inlineEventDataPolicy (int)
--------------------------------------------------------------------------------

Specifies how inline execution event data should be saved in the backend

**Default Value**: 

.. code-block:: yaml

  Offload
  

interfaces.SignedURL
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

enabled (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Whether signed urls should even be returned with GetExecutionData, GetNodeExecutionData and GetTaskExecutionData response objects.

**Default Value**: 

.. code-block:: yaml

  "false"
  

durationMinutes (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

signingPrincipal (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Section: scheduler
================================================================================

profilerPort (`config.Port`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  10254
  

eventScheduler (`interfaces.EventSchedulerConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  aws: null
  local: {}
  region: ""
  scheduleNamePrefix: ""
  scheduleRole: ""
  scheme: local
  targetName: ""
  

workflowExecutor (`interfaces.WorkflowExecutorConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  accountId: ""
  aws: null
  local:
    adminRateLimit:
      burst: 10
      tps: 100
  region: ""
  scheduleQueueName: ""
  scheme: local
  

reconnectAttempts (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

reconnectDelaySeconds (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

config.Port
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

port (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "10254"
  

interfaces.EventSchedulerConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

scheme (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  local
  

region (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

scheduleRole (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

targetName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

scheduleNamePrefix (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

aws (interfaces.AWSSchedulerConfig)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

local (`interfaces.FlyteSchedulerConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  {}
  

interfaces.FlyteSchedulerConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

interfaces.WorkflowExecutorConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

scheme (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  local
  

region (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

scheduleQueueName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

accountId (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

aws (interfaces.AWSWorkflowExecutorConfig)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

local (`interfaces.FlyteWorkflowExecutorConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  adminRateLimit:
    burst: 10
    tps: 100
  

interfaces.FlyteWorkflowExecutorConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

adminRateLimit (`interfaces.AdminRateLimit`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  burst: 10
  tps: 100
  

interfaces.AdminRateLimit
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

tps (float64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "100"
  

burst (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "10"
  

Section: storage
================================================================================

type (string)
--------------------------------------------------------------------------------

Sets the type of storage to configure [s3/minio/local/mem/stow].

**Default Value**: 

.. code-block:: yaml

  s3
  

connection (`storage.ConnectionConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  access-key: ""
  auth-type: iam
  disable-ssl: false
  endpoint: ""
  region: us-east-1
  secret-key: ""
  

stow (`storage.StowConfig`_)
--------------------------------------------------------------------------------

Storage config for stow backend.

**Default Value**: 

.. code-block:: yaml

  {}
  

container (string)
--------------------------------------------------------------------------------

Initial container (in s3 a bucket) to create -if it doesn't exist-.'

**Default Value**: 

.. code-block:: yaml

  ""
  

enable-multicontainer (bool)
--------------------------------------------------------------------------------

If this is true, then the container argument is overlooked and redundant. This config will automatically open new connections to new containers/buckets as they are encountered

**Default Value**: 

.. code-block:: yaml

  "false"
  

cache (`storage.CachingConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  max_size_mbs: 0
  target_gc_percent: 0
  

limits (`storage.LimitsConfig`_)
--------------------------------------------------------------------------------

Sets limits for stores.

**Default Value**: 

.. code-block:: yaml

  maxDownloadMBs: 2
  

defaultHttpClient (`storage.HTTPClientConfig`_)
--------------------------------------------------------------------------------

Sets the default http client config.

**Default Value**: 

.. code-block:: yaml

  headers: null
  timeout: 0s
  

storage.CachingConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

max_size_mbs (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum size of the cache where the Blob store data is cached in-memory. If not specified or set to 0, cache is not used

**Default Value**: 

.. code-block:: yaml

  "0"
  

target_gc_percent (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Sets the garbage collection target percentage.

**Default Value**: 

.. code-block:: yaml

  "0"
  

storage.ConnectionConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

endpoint (`config.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

URL for storage client to connect to.

**Default Value**: 

.. code-block:: yaml

  ""
  

auth-type (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Auth Type to use [iam,accesskey].

**Default Value**: 

.. code-block:: yaml

  iam
  

access-key (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Access key to use. Only required when authtype is set to accesskey.

**Default Value**: 

.. code-block:: yaml

  ""
  

secret-key (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Secret to use when accesskey is set.

**Default Value**: 

.. code-block:: yaml

  ""
  

region (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Region to connect to.

**Default Value**: 

.. code-block:: yaml

  us-east-1
  

disable-ssl (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Disables SSL connection. Should only be used for development.

**Default Value**: 

.. code-block:: yaml

  "false"
  

storage.HTTPClientConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

headers (map[string][]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

timeout (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Sets time out on the http client.

**Default Value**: 

.. code-block:: yaml

  0s
  

storage.LimitsConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

maxDownloadMBs (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum allowed download size (in MBs) per call.

**Default Value**: 

.. code-block:: yaml

  "2"
  

storage.StowConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

kind (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Kind of Stow backend to use. Refer to github/graymeta/stow

**Default Value**: 

.. code-block:: yaml

  ""
  

config (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Configuration for stow backend. Refer to github/graymeta/stow

**Default Value**: 

.. code-block:: yaml

  {}
  

Section: task_resources
================================================================================

defaults (`interfaces.TaskResourceSet`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  cpu: "2"
  ephemeralStorage: "0"
  gpu: "0"
  memory: 200Mi
  storage: "0"
  

limits (`interfaces.TaskResourceSet`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  cpu: "2"
  ephemeralStorage: "0"
  gpu: "1"
  memory: 1Gi
  storage: "0"
  

interfaces.TaskResourceSet
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

cpu (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "2"
  

gpu (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

memory (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  200Mi
  

storage (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

ephemeralStorage (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

resource.Quantity
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

i (`resource.int64Amount`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  {}
  

d (`resource.infDecAmount`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  <nil>
  

s (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "2"
  

Format (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  DecimalSI
  

resource.infDecAmount
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Dec (inf.Dec)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

resource.int64Amount
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

value (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "2"
  

scale (int32)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

