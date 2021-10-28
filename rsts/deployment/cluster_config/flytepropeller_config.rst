.. _flytepropeller-config-specification:

#########################################
Flyte Propeller Configuration
#########################################

admin
------------------------------------
+------------------------+----------------+--------------------------------+
|          NAME          |      TYPE      |          DESCRIPTION           |
+------------------------+----------------+--------------------------------+
| endpoint               | URL_           | For admin types, specify where |
|                        |                | the uri of the service is      |
|                        |                | located.                       |
+------------------------+----------------+--------------------------------+
| insecure               | bool           | Use insecure connection.       |
+------------------------+----------------+--------------------------------+
| insecureSkipVerify     | bool           | InsecureSkipVerify controls    |
|                        |                | whether a client verifies the  |
|                        |                | server's certificate chain and |
|                        |                | host name. Caution : shouldn't |
|                        |                | be use for production          |
|                        |                | usecases'                      |
+------------------------+----------------+--------------------------------+
| maxBackoffDelay        | Duration_      | Max delay for grpc backoff     |
+------------------------+----------------+--------------------------------+
| perRetryTimeout        | Duration_      | gRPC per retry timeout         |
+------------------------+----------------+--------------------------------+
| maxRetries             | int            | Max number of gRPC retries     |
+------------------------+----------------+--------------------------------+
| authType               | admin.AuthType | Type of OAuth2 flow used for   |
|                        |                | communicating with admin.      |
+------------------------+----------------+--------------------------------+
| useAuth                | bool           | Deprecated: Auth will be       |
|                        |                | enabled/disabled based on      |
|                        |                | admin's dynamically discovered |
|                        |                | information.                   |
+------------------------+----------------+--------------------------------+
| clientId               | string         | Client ID                      |
+------------------------+----------------+--------------------------------+
| clientSecretLocation   | string         | File containing the client     |
|                        |                | secret                         |
+------------------------+----------------+--------------------------------+
| scopes                 | []string       | List of scopes to request      |
+------------------------+----------------+--------------------------------+
| authorizationServerUrl | string         | This is the URL to your IdP's  |
|                        |                | authorization server. It'll    |
|                        |                | default to Endpoint            |
+------------------------+----------------+--------------------------------+
| tokenUrl               | string         | OPTIONAL: Your IdP's token     |
|                        |                | endpoint. It'll be discovered  |
|                        |                | from flyte admin's OAuth       |
|                        |                | Metadata endpoint if not       |
|                        |                | provided.                      |
+------------------------+----------------+--------------------------------+
| authorizationHeader    | string         | Custom metadata header to pass |
|                        |                | JWT                            |
+------------------------+----------------+--------------------------------+
| pkceConfig             | Config_        | Config for Pkce authentication |
|                        |                | flow.                          |
+------------------------+----------------+--------------------------------+

Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+-------------+-----------+-------------+
|    NAME     |   TYPE    | DESCRIPTION |
+-------------+-----------+-------------+
| timeout     | Duration_ |             |
+-------------+-----------+-------------+
| refreshTime | Duration_ |             |
+-------------+-----------+-------------+

Duration
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+----------+---------------+-------------+
|   NAME   |     TYPE      | DESCRIPTION |
+----------+---------------+-------------+
| Duration | time.Duration |             |
+----------+---------------+-------------+

URL
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+------+------+-------------+
| NAME | TYPE | DESCRIPTION |
+------+------+-------------+
| URL  | URL_ |             |
+------+------+-------------+

catalog-cache
------------------------------------
+---------------+-----------+--------------------------------+
|     NAME      |   TYPE    |          DESCRIPTION           |
+---------------+-----------+--------------------------------+
| type          | string    |  Catalog Implementation to use |
+---------------+-----------+--------------------------------+
| endpoint      | string    |  Endpoint for catalog service  |
+---------------+-----------+--------------------------------+
| insecure      | bool      |  Use insecure grpc connection  |
+---------------+-----------+--------------------------------+
| max-cache-age | Duration_ |  Cache entries past this age   |
|               |           | will incur cache miss. 0 means |
|               |           | cache never expires            |
+---------------+-----------+--------------------------------+

event
------------------------------------
+-----------+--------+--------------------------------+
|   NAME    |  TYPE  |          DESCRIPTION           |
+-----------+--------+--------------------------------+
| type      | string | Sets the type of EventSink to  |
|           |        | configure [log/admin/file].    |
+-----------+--------+--------------------------------+
| file-path | string | For file types, specify where  |
|           |        | the file should be located.    |
+-----------+--------+--------------------------------+
| rate      | int64  | Max rate at which events can   |
|           |        | be recorded per second.        |
+-----------+--------+--------------------------------+
| capacity  | int    | The max bucket size for event  |
|           |        | recording tokens.              |
+-----------+--------+--------------------------------+

logger
------------------------------------
+-------------+------------------+--------------------------------+
|    NAME     |       TYPE       |          DESCRIPTION           |
+-------------+------------------+--------------------------------+
| show-source | bool             | Includes source code location  |
|             |                  | in logs.                       |
+-------------+------------------+--------------------------------+
| mute        | bool             | Mutes all logs regardless      |
|             |                  | of severity. Intended for      |
|             |                  | benchmarks/tests only.         |
+-------------+------------------+--------------------------------+
| level       | int              | Sets the minimum logging       |
|             |                  | level.                         |
+-------------+------------------+--------------------------------+
| formatter   | FormatterConfig_ | Sets logging format.           |
+-------------+------------------+--------------------------------+

FormatterConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+------+--------+---------------------------+
| NAME |  TYPE  |        DESCRIPTION        |
+------+--------+---------------------------+
| type | string | Sets logging format type. |
+------+--------+---------------------------+

plugins
------------------------------------
+-----------------+----------+--------------------------------+
|      NAME       |   TYPE   |          DESCRIPTION           |
+-----------------+----------+--------------------------------+
| enabled-plugins | []string | List of enabled plugins,       |
|                 |          | default value is to enable all |
|                 |          | plugins.                       |
+-----------------+----------+--------------------------------+

propeller
------------------------------------
+--------------------------+-----------------------+--------------------------------+
|           NAME           |         TYPE          |          DESCRIPTION           |
+--------------------------+-----------------------+--------------------------------+
| kube-config              | string                | Path to kubernetes client      |
|                          |                       | config file.                   |
+--------------------------+-----------------------+--------------------------------+
| master                   | string                |                                |
+--------------------------+-----------------------+--------------------------------+
| workers                  | int                   | Number of threads to process   |
|                          |                       | workflows                      |
+--------------------------+-----------------------+--------------------------------+
| workflow-reeval-duration | Duration_             | Frequency of re-evaluating     |
|                          |                       | workflows                      |
+--------------------------+-----------------------+--------------------------------+
| downstream-eval-duration | Duration_             | Frequency of re-evaluating     |
|                          |                       | downstream tasks               |
+--------------------------+-----------------------+--------------------------------+
| limit-namespace          | string                | Namespaces to watch for this   |
|                          |                       | propeller                      |
+--------------------------+-----------------------+--------------------------------+
| prof-port                | Port_                 | Profiler port                  |
+--------------------------+-----------------------+--------------------------------+
| metadata-prefix          | string                | MetadataPrefix should be       |
|                          |                       | used if all the metadata       |
|                          |                       | for Flyte executions should    |
|                          |                       | be stored under a specific     |
|                          |                       | prefix in CloudStorage. If not |
|                          |                       | specified, the data will be    |
|                          |                       | stored in the base container   |
|                          |                       | directly.                      |
+--------------------------+-----------------------+--------------------------------+
| rawoutput-prefix         | string                | a fully qualified              |
|                          |                       | storage path of the form       |
|                          |                       | s3://flyte/abc/..., where      |
|                          |                       | all data sandboxes should be   |
|                          |                       | stored.                        |
+--------------------------+-----------------------+--------------------------------+
| queue                    | CompositeQueueConfig_ | Workflow workqueue             |
|                          |                       | configuration, affects the way |
|                          |                       | the work is consumed from the  |
|                          |                       | queue.                         |
+--------------------------+-----------------------+--------------------------------+
| metrics-prefix           | string                | An optional prefix for all     |
|                          |                       | published metrics.             |
+--------------------------+-----------------------+--------------------------------+
| enable-admin-launcher    | bool                  | Enable remote Workflow         |
|                          |                       | launcher to Admin              |
+--------------------------+-----------------------+--------------------------------+
| max-workflow-retries     | int                   | Maximum number of retries per  |
|                          |                       | workflow                       |
+--------------------------+-----------------------+--------------------------------+
| max-ttl-hours            | int                   | Maximum number of hours a      |
|                          |                       | completed workflow should be   |
|                          |                       | retained. Number between 1-23  |
|                          |                       | hours                          |
+--------------------------+-----------------------+--------------------------------+
| gc-interval              | Duration_             | Run periodic GC every 30       |
|                          |                       | minutes                        |
+--------------------------+-----------------------+--------------------------------+
| leader-election          | LeaderElectionConfig_ | Config for leader election.    |
+--------------------------+-----------------------+--------------------------------+
| publish-k8s-events       | bool                  | Enable events publishing to    |
|                          |                       | K8s events API.                |
+--------------------------+-----------------------+--------------------------------+
| max-output-size-bytes    | int64                 | Maximum size of outputs per    |
|                          |                       | task                           |
+--------------------------+-----------------------+--------------------------------+
| kube-client-config       | KubeClientConfig_     | Configuration to control the   |
|                          |                       | Kubernetes client              |
+--------------------------+-----------------------+--------------------------------+
| node-config              | NodeConfig_           | config for a workflow node     |
+--------------------------+-----------------------+--------------------------------+
| max-streak-length        | int                   | Maximum number of consecutive  |
|                          |                       | rounds that one propeller      |
|                          |                       | worker can use for one         |
|                          |                       | workflow - >1 => turbo-mode is |
|                          |                       | enabled.                       |
+--------------------------+-----------------------+--------------------------------+
| event-config             | EventConfig_          | Configures execution event     |
|                          |                       | behavior.                      |
+--------------------------+-----------------------+--------------------------------+

EventConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+------------------------------+--------+--------------------------------+
|             NAME             |  TYPE  |          DESCRIPTION           |
+------------------------------+--------+--------------------------------+
| raw-output-policy            | string | How output data should be      |
|                              |        | passed along in execution      |
|                              |        | events.                        |
+------------------------------+--------+--------------------------------+
| fallback-to-output-reference | bool   | Whether output data should be  |
|                              |        | sent by reference when it is   |
|                              |        | too large to be sent inline in |
|                              |        | execution events.              |
+------------------------------+--------+--------------------------------+

NodeConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+----------------------------------+-------------------+--------------------------------+
|               NAME               |       TYPE        |          DESCRIPTION           |
+----------------------------------+-------------------+--------------------------------+
| default-deadlines                | DefaultDeadlines_ | Default value for timeouts     |
+----------------------------------+-------------------+--------------------------------+
| max-node-retries-system-failures | int64             | Maximum number of retries per  |
|                                  |                   | node for node failure due to   |
|                                  |                   | infra issues                   |
+----------------------------------+-------------------+--------------------------------+
| interruptible-failure-threshold  | int64             | number of failures for a       |
|                                  |                   | node to be still considered    |
|                                  |                   | interruptible'                 |
+----------------------------------+-------------------+--------------------------------+

DefaultDeadlines
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+--------------------------+-----------+--------------------------------+
|           NAME           |   TYPE    |          DESCRIPTION           |
+--------------------------+-----------+--------------------------------+
| node-execution-deadline  | Duration_ | Default value of node          |
|                          |           | execution timeout              |
+--------------------------+-----------+--------------------------------+
| node-active-deadline     | Duration_ | Default value of node timeout  |
+--------------------------+-----------+--------------------------------+
| workflow-active-deadline | Duration_ | Default value of workflow      |
|                          |           | timeout                        |
+--------------------------+-----------+--------------------------------+

KubeClientConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+---------+-----------+--------------------------------+
|  NAME   |   TYPE    |          DESCRIPTION           |
+---------+-----------+--------------------------------+
| qps     | float32   | Max QPS to the master for      |
|         |           | requests to KubeAPI. 0         |
|         |           | defaults to 5.                 |
+---------+-----------+--------------------------------+
| burst   | int       | Max burst rate for throttle. 0 |
|         |           | defaults to 10                 |
+---------+-----------+--------------------------------+
| timeout | Duration_ | Max duration allowed for       |
|         |           | every request to KubeAPI       |
|         |           | before giving up. 0 implies no |
|         |           | timeout.                       |
+---------+-----------+--------------------------------+

LeaderElectionConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+-----------------+-----------------+--------------------------------+
|      NAME       |      TYPE       |          DESCRIPTION           |
+-----------------+-----------------+--------------------------------+
| enabled         | bool            | Enables/Disables leader        |
|                 |                 | election.                      |
+-----------------+-----------------+--------------------------------+
| lock-config-map | NamespacedName_ | ConfigMap namespace/name to    |
|                 |                 | use for resource lock.         |
+-----------------+-----------------+--------------------------------+
| lease-duration  | Duration_       | Duration that non-leader       |
|                 |                 | candidates will wait to force  |
|                 |                 | acquire leadership. This is    |
|                 |                 | measured against time of last  |
|                 |                 | observed ack.                  |
+-----------------+-----------------+--------------------------------+
| renew-deadline  | Duration_       | Duration that the acting       |
|                 |                 | master will retry refreshing   |
|                 |                 | leadership before giving up.   |
+-----------------+-----------------+--------------------------------+
| retry-period    | Duration_       | Duration the LeaderElector     |
|                 |                 | clients should wait between    |
|                 |                 | tries of actions.              |
+-----------------+-----------------+--------------------------------+

NamespacedName
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+-----------+--------+-------------+
|   NAME    |  TYPE  | DESCRIPTION |
+-----------+--------+-------------+
| Namespace | string |             |
+-----------+--------+-------------+
| Name      | string |             |
+-----------+--------+-------------+

CompositeQueueConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+-------------------+------------------+--------------------------------+
|       NAME        |       TYPE       |          DESCRIPTION           |
+-------------------+------------------+--------------------------------+
| type              | string           | Type of composite queue to use |
|                   |                  | for the WorkQueue              |
+-------------------+------------------+--------------------------------+
| queue             | WorkqueueConfig_ | Workflow workqueue             |
|                   |                  | configuration, affects the way |
|                   |                  | the work is consumed from the  |
|                   |                  | queue.                         |
+-------------------+------------------+--------------------------------+
| sub-queue         | WorkqueueConfig_ | SubQueue configuration,        |
|                   |                  | affects the way the nodes      |
|                   |                  | cause the top-level Work to be |
|                   |                  | re-evaluated.                  |
+-------------------+------------------+--------------------------------+
| batching-interval | Duration_        | Duration for which downstream  |
|                   |                  | updates are buffered           |
+-------------------+------------------+--------------------------------+
| batch-size        | int              | Number of downstream           |
|                   |                  | triggered top-level objects to |
|                   |                  | re-enqueue every duration. -1  |
|                   |                  | indicates all available.       |
+-------------------+------------------+--------------------------------+

WorkqueueConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+------------+-----------+--------------------------------+
|    NAME    |   TYPE    |          DESCRIPTION           |
+------------+-----------+--------------------------------+
| type       | string    | Type of RateLimiter to use for |
|            |           | the WorkQueue                  |
+------------+-----------+--------------------------------+
| base-delay | Duration_ | base backoff delay for failure |
+------------+-----------+--------------------------------+
| max-delay  | Duration_ | Max backoff delay for failure  |
+------------+-----------+--------------------------------+
| rate       | int64     | Bucket Refill rate per second  |
+------------+-----------+--------------------------------+
| capacity   | int       | Bucket capacity as number of   |
|            |           | items                          |
+------------+-----------+--------------------------------+

Port
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+------+------+-------------+
| NAME | TYPE | DESCRIPTION |
+------+------+-------------+
| port | int  |             |
+------+------+-------------+

secrets
------------------------------------
+----------------+--------+--------------------------------+
|      NAME      |  TYPE  |          DESCRIPTION           |
+----------------+--------+--------------------------------+
| secrets-prefix | string |  Prefix where to look for      |
|                |        | secrets file                   |
+----------------+--------+--------------------------------+
| env-prefix     | string |  Prefix for environment        |
|                |        | variables                      |
+----------------+--------+--------------------------------+

storage
------------------------------------
+-----------------------+-------------------+--------------------------------+
|         NAME          |       TYPE        |          DESCRIPTION           |
+-----------------------+-------------------+--------------------------------+
| type                  | string            | Sets the type of               |
|                       |                   | storage to configure           |
|                       |                   | [s3/minio/local/mem/stow].     |
+-----------------------+-------------------+--------------------------------+
| connection            | ConnectionConfig_ |                                |
+-----------------------+-------------------+--------------------------------+
| stow                  | StowConfig_       | Storage config for stow        |
|                       |                   | backend.                       |
+-----------------------+-------------------+--------------------------------+
| container             | string            | Initial container (in s3       |
|                       |                   | a bucket) to create -if it     |
|                       |                   | doesn't exist-.'               |
+-----------------------+-------------------+--------------------------------+
| enable-multicontainer | bool              | If this is true, then          |
|                       |                   | the container argument is      |
|                       |                   | overlooked and redundant.      |
|                       |                   | This config will automatically |
|                       |                   | open new connections to new    |
|                       |                   | containers/buckets as they are |
|                       |                   | encountered                    |
+-----------------------+-------------------+--------------------------------+
| cache                 | CachingConfig_    |                                |
+-----------------------+-------------------+--------------------------------+
| limits                | LimitsConfig_     | Sets limits for stores.        |
+-----------------------+-------------------+--------------------------------+
| defaultHttpClient     | HTTPClientConfig_ | Sets the default http client   |
|                       |                   | config.                        |
+-----------------------+-------------------+--------------------------------+

HTTPClientConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+---------+---------------------+--------------------------------+
|  NAME   |        TYPE         |          DESCRIPTION           |
+---------+---------------------+--------------------------------+
| headers | map[string][]string | Sets http headers to set on    |
|         |                     | the http client.               |
+---------+---------------------+--------------------------------+
| timeout | Duration_           | Sets time out on the http      |
|         |                     | client.                        |
+---------+---------------------+--------------------------------+

LimitsConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+----------------+-------+--------------------------------+
|      NAME      | TYPE  |          DESCRIPTION           |
+----------------+-------+--------------------------------+
| maxDownloadMBs | int64 | Maximum allowed download size  |
|                |       | (in MBs) per call.             |
+----------------+-------+--------------------------------+

CachingConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+-------------------+------+--------------------------------+
|       NAME        | TYPE |          DESCRIPTION           |
+-------------------+------+--------------------------------+
| max_size_mbs      | int  | Maximum size of the cache      |
|                   |      | where the Blob store data      |
|                   |      | is cached in-memory. If not    |
|                   |      | specified or set to 0, cache   |
|                   |      | is not used                    |
+-------------------+------+--------------------------------+
| target_gc_percent | int  | Sets the garbage collection    |
|                   |      | target percentage.             |
+-------------------+------+--------------------------------+

StowConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+--------+-------------------+--------------------------------+
|  NAME  |       TYPE        |          DESCRIPTION           |
+--------+-------------------+--------------------------------+
| kind   | string            | Kind of Stow backend to use.   |
|        |                   | Refer to github/graymeta/stow  |
+--------+-------------------+--------------------------------+
| config | map[string]string | Configuration for              |
|        |                   | stow backend. Refer to         |
|        |                   | github/graymeta/stow           |
+--------+-------------------+--------------------------------+

ConnectionConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+-------------+--------+--------------------------------+
|    NAME     |  TYPE  |          DESCRIPTION           |
+-------------+--------+--------------------------------+
| endpoint    | URL_   | URL for storage client to      |
|             |        | connect to.                    |
+-------------+--------+--------------------------------+
| auth-type   | string | Auth Type to use               |
|             |        | [iam,accesskey].               |
+-------------+--------+--------------------------------+
| access-key  | string | Access key to use. Only        |
|             |        | required when authtype is set  |
|             |        | to accesskey.                  |
+-------------+--------+--------------------------------+
| secret-key  | string | Secret to use when accesskey   |
|             |        | is set.                        |
+-------------+--------+--------------------------------+
| region      | string | Region to connect to.          |
+-------------+--------+--------------------------------+
| disable-ssl | bool   | Disables SSL connection.       |
|             |        | Should only be used for        |
|             |        | development.                   |
+-------------+--------+--------------------------------+

tasks
------------------------------------
+---------------------------+-------------------+--------------------------------+
|           NAME            |       TYPE        |          DESCRIPTION           |
+---------------------------+-------------------+--------------------------------+
| task-plugins              | TaskPluginConfig_ | Task plugin configuration      |
+---------------------------+-------------------+--------------------------------+
| max-plugin-phase-versions | int32             | Maximum number of plugin       |
|                           |                   | phase versions allowed for one |
|                           |                   | phase.                         |
+---------------------------+-------------------+--------------------------------+
| barrier                   | BarrierConfig_    | Config for Barrier             |
|                           |                   | implementation                 |
+---------------------------+-------------------+--------------------------------+
| backoff                   | BackOffConfig_    | Config for Exponential BackOff |
|                           |                   | implementation                 |
+---------------------------+-------------------+--------------------------------+
| maxLogMessageLength       | int               | Max length of error message.   |
+---------------------------+-------------------+--------------------------------+

BackOffConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+--------------+-----------+--------------------------------+
|     NAME     |   TYPE    |          DESCRIPTION           |
+--------------+-----------+--------------------------------+
| base-second  | int       | The number of seconds          |
|              |           | representing the base duration |
|              |           | of the exponential backoff     |
+--------------+-----------+--------------------------------+
| max-duration | Duration_ | The cap of the backoff         |
|              |           | duration                       |
+--------------+-----------+--------------------------------+

BarrierConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+------------+-----------+--------------------------------+
|    NAME    |   TYPE    |          DESCRIPTION           |
+------------+-----------+--------------------------------+
| enabled    | bool      | Enable Barrier transitions     |
|            |           | using inmemory context         |
+------------+-----------+--------------------------------+
| cache-size | int       | Max number of barrier to       |
|            |           | preserve in memory             |
+------------+-----------+--------------------------------+
| cache-ttl  | Duration_ |  Max duration that a barrier   |
|            |           | would be respected if the      |
|            |           | process is not restarted.      |
|            |           | This should account for time   |
|            |           | required to store the record   |
|            |           | into persistent storage        |
|            |           | (across multiple rounds.       |
+------------+-----------+--------------------------------+

TaskPluginConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+------------------------+-------------------+-------------+
|          NAME          |       TYPE        | DESCRIPTION |
+------------------------+-------------------+-------------+
| enabled-plugins        | []string          | deprecated  |
+------------------------+-------------------+-------------+
| default-for-task-types | map[string]string |             |
+------------------------+-------------------+-------------+

webhook
------------------------------------
+-------------------+--------------------------+--------------------------------+
|       NAME        |           TYPE           |          DESCRIPTION           |
+-------------------+--------------------------+--------------------------------+
| metrics-prefix    | string                   | An optional prefix for all     |
|                   |                          | published metrics.             |
+-------------------+--------------------------+--------------------------------+
| certDir           | string                   | Certificate directory to       |
|                   |                          | use to write generated         |
|                   |                          | certs. Defaults to             |
|                   |                          | /etc/webhook/certs/            |
+-------------------+--------------------------+--------------------------------+
| listenPort        | int                      | The port to use to listen to   |
|                   |                          | webhook calls. Defaults to     |
|                   |                          | 9443                           |
+-------------------+--------------------------+--------------------------------+
| serviceName       | string                   | The name of the webhook        |
|                   |                          | service.                       |
+-------------------+--------------------------+--------------------------------+
| secretName        | string                   | Secret name to write generated |
|                   |                          | certs to.                      |
+-------------------+--------------------------+--------------------------------+
| secretManagerType | config.SecretManagerType | Secret manager type to use     |
|                   |                          | if secrets are not found in    |
|                   |                          | global secrets.                |
+-------------------+--------------------------+--------------------------------+
| awsSecretManager  | AWSSecretManagerConfig_  | AWS Secret Manager config.     |
+-------------------+--------------------------+--------------------------------+

AWSSecretManagerConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+--------------+-----------------------+--------------------------------+
|     NAME     |         TYPE          |          DESCRIPTION           |
+--------------+-----------------------+--------------------------------+
| sidecarImage | string                | Specifies the sidecar docker   |
|              |                       | image to use                   |
+--------------+-----------------------+--------------------------------+
| resources    | ResourceRequirements_ | Specifies resource             |
|              |                       | requirements for the init      |
|              |                       | container.                     |
+--------------+-----------------------+--------------------------------+

ResourceRequirements
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+----------+-----------------+-------------+
|   NAME   |      TYPE       | DESCRIPTION |
+----------+-----------------+-------------+
| limits   | v1.ResourceList |             |
+----------+-----------------+-------------+
| requests | v1.ResourceList |             |
+----------+-----------------+-------------+

