.. _deployment-cluster-config-specification:

Flyte Configuration Specification
---------------------------------------------

Storage Configuration
===============================
Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+-----------------------+--------------------------+--------------------------------+
|         NAME          |           TYPE           |          DESCRIPTION           |
+-----------------------+--------------------------+--------------------------------+
| type                  | string                   | Sets the type of               |
|                       |                          | storage to configure           |
|                       |                          | [s3/minio/local/mem/stow].     |
+-----------------------+--------------------------+--------------------------------+
| connection            | storage.ConnectionConfig |                                |
+-----------------------+--------------------------+--------------------------------+
| stow                  | storage.StowConfig       | Storage config for stow        |
|                       |                          | backend.                       |
+-----------------------+--------------------------+--------------------------------+
| container             | string                   | Initial container (in s3       |
|                       |                          | a bucket) to create -if it     |
|                       |                          | doesn't exist-.'               |
+-----------------------+--------------------------+--------------------------------+
| enable-multicontainer | bool                     | If this is true, then          |
|                       |                          | the container argument is      |
|                       |                          | overlooked and redundant.      |
|                       |                          | This config will automatically |
|                       |                          | open new connections to new    |
|                       |                          | containers/buckets as they are |
|                       |                          | encountered                    |
+-----------------------+--------------------------+--------------------------------+
| cache                 | storage.CachingConfig    |                                |
+-----------------------+--------------------------+--------------------------------+
| limits                | storage.LimitsConfig     | Sets limits for stores.        |
+-----------------------+--------------------------+--------------------------------+
| defaultHttpClient     | storage.HTTPClientConfig | Sets the default http client   |
|                       |                          | config.                        |
+-----------------------+--------------------------+--------------------------------+

HTTPClientConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+---------+---------------------+--------------------------------+
|  NAME   |        TYPE         |          DESCRIPTION           |
+---------+---------------------+--------------------------------+
| headers | map[string][]string | Sets http headers to set on    |
|         |                     | the http client.               |
+---------+---------------------+--------------------------------+
| timeout | config.Duration     | Sets time out on the http      |
|         |                     | client.                        |
+---------+---------------------+--------------------------------+

LimitsConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+----------------+-------+--------------------------------+
|      NAME      | TYPE  |          DESCRIPTION           |
+----------------+-------+--------------------------------+
| maxDownloadMBs | int64 | Maximum allowed download size  |
|                |       | (in MBs) per call.             |
+----------------+-------+--------------------------------+

CachingConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
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
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
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
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+-------------+------------+--------------------------------+
|    NAME     |    TYPE    |          DESCRIPTION           |
+-------------+------------+--------------------------------+
| endpoint    | config.URL | URL for storage client to      |
|             |            | connect to.                    |
+-------------+------------+--------------------------------+
| auth-type   | string     | Auth Type to use               |
|             |            | [iam,accesskey].               |
+-------------+------------+--------------------------------+
| access-key  | string     | Access key to use. Only        |
|             |            | required when authtype is set  |
|             |            | to accesskey.                  |
+-------------+------------+--------------------------------+
| secret-key  | string     | Secret to use when accesskey   |
|             |            | is set.                        |
+-------------+------------+--------------------------------+
| region      | string     | Region to connect to.          |
+-------------+------------+--------------------------------+
| disable-ssl | bool       | Disables SSL connection.       |
|             |            | Should only be used for        |
|             |            | development.                   |
+-------------+------------+--------------------------------+

Logger Configuration
===============================
Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+-------------+------------------------+--------------------------------+
|    NAME     |          TYPE          |          DESCRIPTION           |
+-------------+------------------------+--------------------------------+
| show-source | bool                   | Includes source code location  |
|             |                        | in logs.                       |
+-------------+------------------------+--------------------------------+
| mute        | bool                   | Mutes all logs regardless      |
|             |                        | of severity. Intended for      |
|             |                        | benchmarks/tests only.         |
+-------------+------------------------+--------------------------------+
| level       | int                    | Sets the minimum logging       |
|             |                        | level.                         |
+-------------+------------------------+--------------------------------+
| formatter   | logger.FormatterConfig | Sets logging format.           |
+-------------+------------------------+--------------------------------+

FormatterConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+------+--------+---------------------------+
| NAME |  TYPE  |        DESCRIPTION        |
+------+--------+---------------------------+
| type | string | Sets logging format type. |
+------+--------+---------------------------+

Propeller Configuration
===============================
Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+--------------------------+-----------------------------+--------------------------------+
|           NAME           |            TYPE             |          DESCRIPTION           |
+--------------------------+-----------------------------+--------------------------------+
| kube-config              | string                      | Path to kubernetes client      |
|                          |                             | config file.                   |
+--------------------------+-----------------------------+--------------------------------+
| master                   | string                      |                                |
+--------------------------+-----------------------------+--------------------------------+
| workers                  | int                         | Number of threads to process   |
|                          |                             | workflows                      |
+--------------------------+-----------------------------+--------------------------------+
| workflow-reeval-duration | config.Duration             | Frequency of re-evaluating     |
|                          |                             | workflows                      |
+--------------------------+-----------------------------+--------------------------------+
| downstream-eval-duration | config.Duration             | Frequency of re-evaluating     |
|                          |                             | downstream tasks               |
+--------------------------+-----------------------------+--------------------------------+
| limit-namespace          | string                      | Namespaces to watch for this   |
|                          |                             | propeller                      |
+--------------------------+-----------------------------+--------------------------------+
| prof-port                | config.Port                 | Profiler port                  |
+--------------------------+-----------------------------+--------------------------------+
| metadata-prefix          | string                      | MetadataPrefix should be       |
|                          |                             | used if all the metadata       |
|                          |                             | for Flyte executions should    |
|                          |                             | be stored under a specific     |
|                          |                             | prefix in CloudStorage. If not |
|                          |                             | specified, the data will be    |
|                          |                             | stored in the base container   |
|                          |                             | directly.                      |
+--------------------------+-----------------------------+--------------------------------+
| rawoutput-prefix         | string                      | a fully qualified              |
|                          |                             | storage path of the form       |
|                          |                             | s3://flyte/abc/..., where      |
|                          |                             | all data sandboxes should be   |
|                          |                             | stored.                        |
+--------------------------+-----------------------------+--------------------------------+
| queue                    | config.CompositeQueueConfig | Workflow workqueue             |
|                          |                             | configuration, affects the way |
|                          |                             | the work is consumed from the  |
|                          |                             | queue.                         |
+--------------------------+-----------------------------+--------------------------------+
| metrics-prefix           | string                      | An optional prefix for all     |
|                          |                             | published metrics.             |
+--------------------------+-----------------------------+--------------------------------+
| enable-admin-launcher    | bool                        | Enable remote Workflow         |
|                          |                             | launcher to Admin              |
+--------------------------+-----------------------------+--------------------------------+
| max-workflow-retries     | int                         | Maximum number of retries per  |
|                          |                             | workflow                       |
+--------------------------+-----------------------------+--------------------------------+
| max-ttl-hours            | int                         | Maximum number of hours a      |
|                          |                             | completed workflow should be   |
|                          |                             | retained. Number between 1-23  |
|                          |                             | hours                          |
+--------------------------+-----------------------------+--------------------------------+
| gc-interval              | config.Duration             | Run periodic GC every 30       |
|                          |                             | minutes                        |
+--------------------------+-----------------------------+--------------------------------+
| leader-election          | config.LeaderElectionConfig | Config for leader election.    |
+--------------------------+-----------------------------+--------------------------------+
| publish-k8s-events       | bool                        | Enable events publishing to    |
|                          |                             | K8s events API.                |
+--------------------------+-----------------------------+--------------------------------+
| max-output-size-bytes    | int64                       | Maximum size of outputs per    |
|                          |                             | task                           |
+--------------------------+-----------------------------+--------------------------------+
| kube-client-config       | config.KubeClientConfig     | Configuration to control the   |
|                          |                             | Kubernetes client              |
+--------------------------+-----------------------------+--------------------------------+
| node-config              | config.NodeConfig           | config for a workflow node     |
+--------------------------+-----------------------------+--------------------------------+
| max-streak-length        | int                         | Maximum number of consecutive  |
|                          |                             | rounds that one propeller      |
|                          |                             | worker can use for one         |
|                          |                             | workflow - >1 => turbo-mode is |
|                          |                             | enabled.                       |
+--------------------------+-----------------------------+--------------------------------+
| event-config             | config.EventConfig          | Configures execution event     |
|                          |                             | behavior.                      |
+--------------------------+-----------------------------+--------------------------------+

EventConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
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
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+----------------------------------+-------------------------+--------------------------------+
|               NAME               |          TYPE           |          DESCRIPTION           |
+----------------------------------+-------------------------+--------------------------------+
| default-deadlines                | config.DefaultDeadlines | Default value for timeouts     |
+----------------------------------+-------------------------+--------------------------------+
| max-node-retries-system-failures | int64                   | Maximum number of retries per  |
|                                  |                         | node for node failure due to   |
|                                  |                         | infra issues                   |
+----------------------------------+-------------------------+--------------------------------+
| interruptible-failure-threshold  | int64                   | number of failures for a       |
|                                  |                         | node to be still considered    |
|                                  |                         | interruptible'                 |
+----------------------------------+-------------------------+--------------------------------+

DefaultDeadlines
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+--------------------------+-----------------+--------------------------------+
|           NAME           |      TYPE       |          DESCRIPTION           |
+--------------------------+-----------------+--------------------------------+
| node-execution-deadline  | config.Duration | Default value of node          |
|                          |                 | execution timeout              |
+--------------------------+-----------------+--------------------------------+
| node-active-deadline     | config.Duration | Default value of node timeout  |
+--------------------------+-----------------+--------------------------------+
| workflow-active-deadline | config.Duration | Default value of workflow      |
|                          |                 | timeout                        |
+--------------------------+-----------------+--------------------------------+

KubeClientConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+---------+-----------------+--------------------------------+
|  NAME   |      TYPE       |          DESCRIPTION           |
+---------+-----------------+--------------------------------+
| qps     | float32         | Max QPS to the master for      |
|         |                 | requests to KubeAPI. 0         |
|         |                 | defaults to 5.                 |
+---------+-----------------+--------------------------------+
| burst   | int             | Max burst rate for throttle. 0 |
|         |                 | defaults to 10                 |
+---------+-----------------+--------------------------------+
| timeout | config.Duration | Max duration allowed for       |
|         |                 | every request to KubeAPI       |
|         |                 | before giving up. 0 implies no |
|         |                 | timeout.                       |
+---------+-----------------+--------------------------------+

LeaderElectionConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+-----------------+----------------------+--------------------------------+
|      NAME       |         TYPE         |          DESCRIPTION           |
+-----------------+----------------------+--------------------------------+
| enabled         | bool                 | Enables/Disables leader        |
|                 |                      | election.                      |
+-----------------+----------------------+--------------------------------+
| lock-config-map | types.NamespacedName | ConfigMap namespace/name to    |
|                 |                      | use for resource lock.         |
+-----------------+----------------------+--------------------------------+
| lease-duration  | config.Duration      | Duration that non-leader       |
|                 |                      | candidates will wait to force  |
|                 |                      | acquire leadership. This is    |
|                 |                      | measured against time of last  |
|                 |                      | observed ack.                  |
+-----------------+----------------------+--------------------------------+
| renew-deadline  | config.Duration      | Duration that the acting       |
|                 |                      | master will retry refreshing   |
|                 |                      | leadership before giving up.   |
+-----------------+----------------------+--------------------------------+
| retry-period    | config.Duration      | Duration the LeaderElector     |
|                 |                      | clients should wait between    |
|                 |                      | tries of actions.              |
+-----------------+----------------------+--------------------------------+

CompositeQueueConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+-------------------+------------------------+--------------------------------+
|       NAME        |          TYPE          |          DESCRIPTION           |
+-------------------+------------------------+--------------------------------+
| type              | string                 | Type of composite queue to use |
|                   |                        | for the WorkQueue              |
+-------------------+------------------------+--------------------------------+
| queue             | config.WorkqueueConfig | Workflow workqueue             |
|                   |                        | configuration, affects the way |
|                   |                        | the work is consumed from the  |
|                   |                        | queue.                         |
+-------------------+------------------------+--------------------------------+
| sub-queue         | config.WorkqueueConfig | SubQueue configuration,        |
|                   |                        | affects the way the nodes      |
|                   |                        | cause the top-level Work to be |
|                   |                        | re-evaluated.                  |
+-------------------+------------------------+--------------------------------+
| batching-interval | config.Duration        | Duration for which downstream  |
|                   |                        | updates are buffered           |
+-------------------+------------------------+--------------------------------+
| batch-size        | int                    | Number of downstream           |
|                   |                        | triggered top-level objects to |
|                   |                        | re-enqueue every duration. -1  |
|                   |                        | indicates all available.       |
+-------------------+------------------------+--------------------------------+

WorkqueueConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+------------+-----------------+--------------------------------+
|    NAME    |      TYPE       |          DESCRIPTION           |
+------------+-----------------+--------------------------------+
| type       | string          | Type of RateLimiter to use for |
|            |                 | the WorkQueue                  |
+------------+-----------------+--------------------------------+
| base-delay | config.Duration | base backoff delay for failure |
+------------+-----------------+--------------------------------+
| max-delay  | config.Duration | Max backoff delay for failure  |
+------------+-----------------+--------------------------------+
| rate       | int64           | Bucket Refill rate per second  |
+------------+-----------------+--------------------------------+
| capacity   | int             | Bucket capacity as number of   |
|            |                 | items                          |
+------------+-----------------+--------------------------------+

WorkqueueConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+------------+-----------------+--------------------------------+
|    NAME    |      TYPE       |          DESCRIPTION           |
+------------+-----------------+--------------------------------+
| type       | string          | Type of RateLimiter to use for |
|            |                 | the WorkQueue                  |
+------------+-----------------+--------------------------------+
| base-delay | config.Duration | base backoff delay for failure |
+------------+-----------------+--------------------------------+
| max-delay  | config.Duration | Max backoff delay for failure  |
+------------+-----------------+--------------------------------+
| rate       | int64           | Bucket Refill rate per second  |
+------------+-----------------+--------------------------------+
| capacity   | int             | Bucket capacity as number of   |
|            |                 | items                          |
+------------+-----------------+--------------------------------+

ResourceManager Configuration
===============================
Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+------------------+--------------------+--------------------------------+
|       NAME       |        TYPE        |          DESCRIPTION           |
+------------------+--------------------+--------------------------------+
| type             | string             | Which resource manager to use  |
+------------------+--------------------+--------------------------------+
| resourceMaxQuota | int                | Global limit for concurrent    |
|                  |                    | Qubole queries                 |
+------------------+--------------------+--------------------------------+
| redis            | config.RedisConfig | Config for Redis               |
|                  |                    | resourcemanager.               |
+------------------+--------------------+--------------------------------+

RedisConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
+-------------+----------+--------------------------------+
|    NAME     |   TYPE   |          DESCRIPTION           |
+-------------+----------+--------------------------------+
| hostPaths   | []string | Redis hosts locations.         |
+-------------+----------+--------------------------------+
| primaryName | string   | Redis primary name, fill in    |
|             |          | only if you are connecting to  |
|             |          | a redis sentinel cluster.      |
+-------------+----------+--------------------------------+
| hostPath    | string   | Redis host location            |
+-------------+----------+--------------------------------+
| hostKey     | string   | Key for local Redis access     |
+-------------+----------+--------------------------------+
| maxRetries  | int      | See Redis client options for   |
|             |          | more info                      |
+-------------+----------+--------------------------------+

