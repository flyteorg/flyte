.. _flytepropeller-config-specification:

#########################################
Flyte Propeller Configuration
#########################################

Section: admin
================================================================================

endpoint `config.URL`_
--------------------------------------------------------------------------------

**Description**: For admin types, specify where the uri of the service is located.

**Default Value**: 

.. code-block:: yaml

  ""
  

insecure bool
--------------------------------------------------------------------------------

**Description**: Use insecure connection.

**Default Value**: 

.. code-block:: yaml

  "false"
  

insecureSkipVerify bool
--------------------------------------------------------------------------------

**Description**: InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name. Caution : shouldn't be use for production usecases'

**Default Value**: 

.. code-block:: yaml

  "false"
  

maxBackoffDelay `config.Duration`_
--------------------------------------------------------------------------------

**Description**: Max delay for grpc backoff

**Default Value**: 

.. code-block:: yaml

  8s
  

perRetryTimeout `config.Duration`_
--------------------------------------------------------------------------------

**Description**: gRPC per retry timeout

**Default Value**: 

.. code-block:: yaml

  15s
  

maxRetries int
--------------------------------------------------------------------------------

**Description**: Max number of gRPC retries

**Default Value**: 

.. code-block:: yaml

  "4"
  

authType uint8
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ClientSecret
  

useAuth bool
--------------------------------------------------------------------------------

**Description**: Deprecated: Auth will be enabled/disabled based on admin's dynamically discovered information.

**Default Value**: 

.. code-block:: yaml

  "false"
  

clientId string
--------------------------------------------------------------------------------

**Description**: Client ID

**Default Value**: 

.. code-block:: yaml

  flytepropeller
  

clientSecretLocation string
--------------------------------------------------------------------------------

**Description**: File containing the client secret

**Default Value**: 

.. code-block:: yaml

  /etc/secrets/client_secret
  

scopes []string
--------------------------------------------------------------------------------

**Description**: List of scopes to request

**Default Value**: 

.. code-block:: yaml

  []
  

authorizationServerUrl string
--------------------------------------------------------------------------------

**Description**: This is the URL to your IdP's authorization server. It'll default to Endpoint

**Default Value**: 

.. code-block:: yaml

  ""
  

tokenUrl string
--------------------------------------------------------------------------------

**Description**: OPTIONAL: Your IdP's token endpoint. It'll be discovered from flyte admin's OAuth Metadata endpoint if not provided.

**Default Value**: 

.. code-block:: yaml

  ""
  

authorizationHeader string
--------------------------------------------------------------------------------

**Description**: Custom metadata header to pass JWT

**Default Value**: 

.. code-block:: yaml

  ""
  

pkceConfig `pkce.Config`_
--------------------------------------------------------------------------------

**Description**: Config for Pkce authentication flow.

**Default Value**: 

.. code-block:: yaml

  refreshTime: 5m0s
  timeout: 15s
  

command []string
--------------------------------------------------------------------------------

**Description**: Command for external authentication token generation

**Default Value**: 

.. code-block:: yaml

  null
  

config.Duration
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Duration int64
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  8s
  

config.URL
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

URL `url.URL`_
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

Scheme string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Opaque string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

User url.Userinfo
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

Host string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Path string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

RawPath string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

ForceQuery bool
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "false"
  

RawQuery string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Fragment string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

RawFragment string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

pkce.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

timeout `config.Duration`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  15s
  

refreshTime `config.Duration`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  5m0s
  

Section: catalog-cache
================================================================================

type string
--------------------------------------------------------------------------------

**Description**: Catalog Implementation to use

**Default Value**: 

.. code-block:: yaml

  noop
  

endpoint string
--------------------------------------------------------------------------------

**Description**: Endpoint for catalog service

**Default Value**: 

.. code-block:: yaml

  ""
  

insecure bool
--------------------------------------------------------------------------------

**Description**: Use insecure grpc connection

**Default Value**: 

.. code-block:: yaml

  "false"
  

max-cache-age `config.Duration`_
--------------------------------------------------------------------------------

**Description**: Cache entries past this age will incur cache miss. 0 means cache never expires

**Default Value**: 

.. code-block:: yaml

  0s
  

Section: event
================================================================================

type string
--------------------------------------------------------------------------------

**Description**: Sets the type of EventSink to configure [log/admin/file].

**Default Value**: 

.. code-block:: yaml

  ""
  

file-path string
--------------------------------------------------------------------------------

**Description**: For file types, specify where the file should be located.

**Default Value**: 

.. code-block:: yaml

  ""
  

rate int64
--------------------------------------------------------------------------------

**Description**: Max rate at which events can be recorded per second.

**Default Value**: 

.. code-block:: yaml

  "500"
  

capacity int
--------------------------------------------------------------------------------

**Description**: The max bucket size for event recording tokens.

**Default Value**: 

.. code-block:: yaml

  "1000"
  

Section: logger
================================================================================

show-source bool
--------------------------------------------------------------------------------

**Description**: Includes source code location in logs.

**Default Value**: 

.. code-block:: yaml

  "false"
  

mute bool
--------------------------------------------------------------------------------

**Description**: Mutes all logs regardless of severity. Intended for benchmarks/tests only.

**Default Value**: 

.. code-block:: yaml

  "false"
  

level int
--------------------------------------------------------------------------------

**Description**: Sets the minimum logging level.

**Default Value**: 

.. code-block:: yaml

  "4"
  

formatter `logger.FormatterConfig`_
--------------------------------------------------------------------------------

**Description**: Sets logging format.

**Default Value**: 

.. code-block:: yaml

  type: json
  

logger.FormatterConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

type string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Sets logging format type.

**Default Value**: 

.. code-block:: yaml

  json
  

Section: plugins
================================================================================

enabled-plugins []string
--------------------------------------------------------------------------------

**Description**: List of enabled plugins, default value is to enable all plugins.

**Default Value**: 

.. code-block:: yaml

  - '*'
  

athena `athena.Config`_
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  defaultCatalog: AwsDataCatalog
  defaultWorkGroup: primary
  resourceConstraints:
    NamespaceScopeResourceConstraint:
      Value: 50
    ProjectScopeResourceConstraint:
      Value: 100
  webApi:
    caching:
      maxSystemFailures: 5
      resyncInterval: 30s
      size: 500000
      workers: 10
    readRateLimiter:
      burst: 100
      qps: 10
    resourceMeta: null
    resourceQuotas:
      default: 1000
    writeRateLimiter:
      burst: 100
      qps: 10
  

aws `aws.Config`_
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  accountId: ""
  logLevel: 0
  region: us-east-2
  retries: 3
  

catalogcache `catalog.Config`_
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  reader:
    maxItems: 1000
    maxRetries: 3
    workers: 10
  writer:
    maxItems: 1000
    maxRetries: 3
    workers: 10
  

k8s `config.K8sPluginConfig`_
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  co-pilot:
    cpu: 500m
    default-input-path: /var/flyte/inputs
    default-output-path: /var/flyte/outputs
    image: cr.flyte.org/flyteorg/flytecopilot:v0.0.9
    input-vol-name: flyte-inputs
    memory: 128Mi
    name: flyte-copilot-
    output-vol-name: flyte-outputs
    start-timeout: 1m0s
    storage: ""
  create-container-error-grace-period: 3m0s
  default-annotations:
    cluster-autoscaler.kubernetes.io/safe-to-evict: "false"
  default-cpus: "1"
  default-env-vars: null
  default-env-vars-from-env: null
  default-labels: null
  default-memory: 1Gi
  default-node-selector: null
  default-tolerations: null
  delete-resource-on-finalize: false
  gpu-resource-name: nvidia.com/gpu
  inject-finalizer: false
  interruptible-node-selector: null
  interruptible-node-selector-requirement: null
  interruptible-tolerations: null
  non-interruptible-node-selector-requirement: null
  resource-tolerations: null
  scheduler-name: ""
  

k8s-array `k8s.Config`_
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ErrorAssembler:
    maxItems: 100000
    maxRetries: 5
    workers: 10
  OutputAssembler:
    maxItems: 100000
    maxRetries: 5
    workers: 10
  logs:
    config:
      cloudwatch-enabled: false
      cloudwatch-log-group: ""
      cloudwatch-region: ""
      cloudwatch-template-uri: ""
      gcp-project: ""
      kubernetes-enabled: false
      kubernetes-template-uri: ""
      kubernetes-url: ""
      stackdriver-enabled: false
      stackdriver-logresourcename: ""
      stackdriver-template-uri: ""
      templates: null
  maxArrayJobSize: 5000
  maxErrorLength: 1000
  namespaceTemplate: ""
  node-selector: null
  remoteClusterConfig:
    auth:
      certPath: ""
      tokenPath: ""
      type: ""
    enabled: false
    endpoint: ""
    name: ""
  resourceConfig:
    limit: 0
    primaryLabel: ""
  scheduler: ""
  tolerations: null
  

logs `logs.LogConfig`_
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  cloudwatch-enabled: false
  cloudwatch-log-group: ""
  cloudwatch-region: ""
  cloudwatch-template-uri: ""
  gcp-project: ""
  kubernetes-enabled: false
  kubernetes-template-uri: ""
  kubernetes-url: ""
  stackdriver-enabled: false
  stackdriver-logresourcename: ""
  stackdriver-template-uri: ""
  templates: null
  

qubole `config.Config`_
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  analyzeLinkPath: /v2/analyze
  clusterConfigs:
  - labels:
    - default
    limit: 100
    namespaceScopeQuotaProportionCap: 0.7
    primaryLabel: default
    projectScopeQuotaProportionCap: 0.7
  commandApiPath: /api/v1.2/commands/
  defaultClusterLabel: default
  destinationClusterConfigs: []
  endpoint: https://wellness.qubole.com
  lruCacheSize: 2000
  quboleTokenKey: FLYTE_QUBOLE_CLIENT_TOKEN
  workers: 15
  

sagemaker `config.Config (sagemaker)`_
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  prebuiltAlgorithms:
  - name: xgboost
    regionalConfigs:
    - region: us-east-1
      versionConfigs:
      - image: 683313688378.dkr.ecr.us-east-1.amazonaws.com/sagemaker-xgboost:0.90-2-cpu-py3
        version: "0.90"
  region: us-east-1
  roleAnnotationKey: ""
  roleArn: default_role
  

snowflake `snowflake.Config`_
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  defaultWarehouse: COMPUTE_WH
  resourceConstraints:
    NamespaceScopeResourceConstraint:
      Value: 50
    ProjectScopeResourceConstraint:
      Value: 100
  snowflakeTokenKey: FLYTE_SNOWFLAKE_CLIENT_TOKEN
  webApi:
    caching:
      maxSystemFailures: 5
      resyncInterval: 30s
      size: 500000
      workers: 10
    readRateLimiter:
      burst: 100
      qps: 10
    resourceMeta: null
    resourceQuotas:
      default: 1000
    writeRateLimiter:
      burst: 100
      qps: 10
  

spark `spark.Config`_
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  features: null
  logs:
    all-user:
      cloudwatch-enabled: false
      cloudwatch-log-group: ""
      cloudwatch-region: ""
      cloudwatch-template-uri: ""
      gcp-project: ""
      kubernetes-enabled: false
      kubernetes-template-uri: ""
      kubernetes-url: ""
      stackdriver-enabled: false
      stackdriver-logresourcename: ""
      stackdriver-template-uri: ""
      templates: null
    mixed:
      cloudwatch-enabled: false
      cloudwatch-log-group: ""
      cloudwatch-region: ""
      cloudwatch-template-uri: ""
      gcp-project: ""
      kubernetes-enabled: true
      kubernetes-template-uri: http://localhost:30082/#!/log/{{ .namespace }}/{{ .podName
        }}/pod?namespace={{ .namespace }}
      kubernetes-url: ""
      stackdriver-enabled: false
      stackdriver-logresourcename: ""
      stackdriver-template-uri: ""
      templates: null
    system:
      cloudwatch-enabled: false
      cloudwatch-log-group: ""
      cloudwatch-region: ""
      cloudwatch-template-uri: ""
      gcp-project: ""
      kubernetes-enabled: false
      kubernetes-template-uri: ""
      kubernetes-url: ""
      stackdriver-enabled: false
      stackdriver-logresourcename: ""
      stackdriver-template-uri: ""
      templates: null
    user:
      cloudwatch-enabled: false
      cloudwatch-log-group: ""
      cloudwatch-region: ""
      cloudwatch-template-uri: ""
      gcp-project: ""
      kubernetes-enabled: false
      kubernetes-template-uri: ""
      kubernetes-url: ""
      stackdriver-enabled: false
      stackdriver-logresourcename: ""
      stackdriver-template-uri: ""
      templates: null
  spark-config-default: null
  spark-history-server-url: ""
  

athena.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

webApi `webapi.PluginConfig`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines config for the base WebAPI plugin.

**Default Value**: 

.. code-block:: yaml

  caching:
    maxSystemFailures: 5
    resyncInterval: 30s
    size: 500000
    workers: 10
  readRateLimiter:
    burst: 100
    qps: 10
  resourceMeta: null
  resourceQuotas:
    default: 1000
  writeRateLimiter:
    burst: 100
    qps: 10
  

resourceConstraints `core.ResourceConstraintsSpec`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  NamespaceScopeResourceConstraint:
    Value: 50
  ProjectScopeResourceConstraint:
    Value: 100
  

defaultWorkGroup string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines the default workgroup to use when running on Athena unless overwritten by the task.

**Default Value**: 

.. code-block:: yaml

  primary
  

defaultCatalog string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines the default catalog to use when running on Athena unless overwritten by the task.

**Default Value**: 

.. code-block:: yaml

  AwsDataCatalog
  

core.ResourceConstraintsSpec
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

ProjectScopeResourceConstraint `core.ResourceConstraint`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  Value: 100
  

NamespaceScopeResourceConstraint `core.ResourceConstraint`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  Value: 50
  

core.ResourceConstraint
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Value int64
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "100"
  

webapi.PluginConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

resourceQuotas webapi.ResourceQuotas
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  default: 1000
  

readRateLimiter `webapi.RateLimiterConfig`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines rate limiter properties for read actions (e.g. retrieve status).

**Default Value**: 

.. code-block:: yaml

  burst: 100
  qps: 10
  

writeRateLimiter `webapi.RateLimiterConfig`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines rate limiter properties for write actions.

**Default Value**: 

.. code-block:: yaml

  burst: 100
  qps: 10
  

caching `webapi.CachingConfig`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines caching characteristics.

**Default Value**: 

.. code-block:: yaml

  maxSystemFailures: 5
  resyncInterval: 30s
  size: 500000
  workers: 10
  

resourceMeta interface
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  <nil>
  

webapi.CachingConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

size int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines the maximum number of items to cache.

**Default Value**: 

.. code-block:: yaml

  "500000"
  

resyncInterval `config.Duration`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines the sync interval.

**Default Value**: 

.. code-block:: yaml

  30s
  

workers int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines the number of workers to start up to process items.

**Default Value**: 

.. code-block:: yaml

  "10"
  

maxSystemFailures int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines the number of failures to fetch a task before failing the task.

**Default Value**: 

.. code-block:: yaml

  "5"
  

webapi.RateLimiterConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

qps int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines the max rate of calls per second.

**Default Value**: 

.. code-block:: yaml

  "10"
  

burst int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines the maximum burst size.

**Default Value**: 

.. code-block:: yaml

  "100"
  

aws.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

region string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: AWS Region to connect to.

**Default Value**: 

.. code-block:: yaml

  us-east-2
  

accountId string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: AWS Account Identifier.

**Default Value**: 

.. code-block:: yaml

  ""
  

retries int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Number of retries.

**Default Value**: 

.. code-block:: yaml

  "3"
  

logLevel uint64
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

catalog.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

reader `workqueue.Config`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Catalog reader workqueue config. Make sure the index cache must be big enough to accommodate the biggest array task allowed to run on the system.

**Default Value**: 

.. code-block:: yaml

  maxItems: 1000
  maxRetries: 3
  workers: 10
  

writer `workqueue.Config`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Catalog writer workqueue config. Make sure the index cache must be big enough to accommodate the biggest array task allowed to run on the system.

**Default Value**: 

.. code-block:: yaml

  maxItems: 1000
  maxRetries: 3
  workers: 10
  

workqueue.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

workers int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Number of concurrent workers to start processing the queue.

**Default Value**: 

.. code-block:: yaml

  "10"
  

maxRetries int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Maximum number of retries per item.

**Default Value**: 

.. code-block:: yaml

  "3"
  

maxItems int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Maximum number of entries to keep in the index.

**Default Value**: 

.. code-block:: yaml

  "1000"
  

config.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

endpoint `config.URL`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Endpoint for qubole to use

**Default Value**: 

.. code-block:: yaml

  https://wellness.qubole.com
  

commandApiPath `config.URL`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: API Path where commands can be launched on Qubole. Should be a valid url.

**Default Value**: 

.. code-block:: yaml

  /api/v1.2/commands/
  

analyzeLinkPath `config.URL`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: URL path where queries can be visualized on qubole website. Should be a valid url.

**Default Value**: 

.. code-block:: yaml

  /v2/analyze
  

quboleTokenKey string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Name of the key where to find Qubole token in the secret manager.

**Default Value**: 

.. code-block:: yaml

  FLYTE_QUBOLE_CLIENT_TOKEN
  

lruCacheSize int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Size of the AutoRefreshCache

**Default Value**: 

.. code-block:: yaml

  "2000"
  

workers int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Number of parallel workers to refresh the cache

**Default Value**: 

.. code-block:: yaml

  "15"
  

defaultClusterLabel string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: The default cluster label. This will be used if label is not specified on the hive job.

**Default Value**: 

.. code-block:: yaml

  default
  

clusterConfigs []config.ClusterConfig
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  - labels:
    - default
    limit: 100
    namespaceScopeQuotaProportionCap: 0.7
    primaryLabel: default
    projectScopeQuotaProportionCap: 0.7
  

destinationClusterConfigs []config.DestinationClusterConfig
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  []
  

config.Config (sagemaker)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

roleArn string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: The role the SageMaker plugin uses to communicate with the SageMaker service

**Default Value**: 

.. code-block:: yaml

  default_role
  

region string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: The AWS region the SageMaker plugin communicates to

**Default Value**: 

.. code-block:: yaml

  us-east-1
  

roleAnnotationKey string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Map key to use to lookup role from task annotations.

**Default Value**: 

.. code-block:: yaml

  ""
  

prebuiltAlgorithms []config.PrebuiltAlgorithmConfig
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  - name: xgboost
    regionalConfigs:
    - region: us-east-1
      versionConfigs:
      - image: 683313688378.dkr.ecr.us-east-1.amazonaws.com/sagemaker-xgboost:0.90-2-cpu-py3
        version: "0.90"
  

config.K8sPluginConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

inject-finalizer bool
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Instructs the plugin to inject a finalizer on startTask and remove it on task termination.

**Default Value**: 

.. code-block:: yaml

  "false"
  

default-annotations map[string]string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  cluster-autoscaler.kubernetes.io/safe-to-evict: "false"
  

default-labels map[string]string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-env-vars map[string]string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-env-vars-from-env map[string]string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-cpus `resource.Quantity`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines a default value for cpu for containers if not specified.

**Default Value**: 

.. code-block:: yaml

  "1"
  

default-memory `resource.Quantity`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines a default value for memory for containers if not specified.

**Default Value**: 

.. code-block:: yaml

  1Gi
  

default-tolerations []v1.Toleration
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-node-selector map[string]string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-affinity v1.Affinity
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

scheduler-name string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines scheduler name.

**Default Value**: 

.. code-block:: yaml

  ""
  

interruptible-tolerations []v1.Toleration
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

interruptible-node-selector map[string]string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

interruptible-node-selector-requirement v1.NodeSelectorRequirement
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

non-interruptible-node-selector-requirement v1.NodeSelectorRequirement
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

resource-tolerations map[v1.ResourceName][]v1.Toleration
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

co-pilot `config.FlyteCoPilotConfig`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Co-Pilot Configuration

**Default Value**: 

.. code-block:: yaml

  cpu: 500m
  default-input-path: /var/flyte/inputs
  default-output-path: /var/flyte/outputs
  image: cr.flyte.org/flyteorg/flytecopilot:v0.0.9
  input-vol-name: flyte-inputs
  memory: 128Mi
  name: flyte-copilot-
  output-vol-name: flyte-outputs
  start-timeout: 1m0s
  storage: ""
  

delete-resource-on-finalize bool
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Instructs the system to delete the resource on finalize. This ensures that no resources are kept around (potentially consuming cluster resources). This, however, will cause k8s log links to expire as soon as the resource is finalized.

**Default Value**: 

.. code-block:: yaml

  "false"
  

create-container-error-grace-period `config.Duration`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  3m0s
  

gpu-resource-name string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: The name of the GPU resource to use when the task resource requests GPUs.

**Default Value**: 

.. code-block:: yaml

  nvidia.com/gpu
  

config.FlyteCoPilotConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

name string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Flyte co-pilot sidecar container name prefix. (additional bits will be added after this)

**Default Value**: 

.. code-block:: yaml

  flyte-copilot-
  

image string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Flyte co-pilot Docker Image FQN

**Default Value**: 

.. code-block:: yaml

  cr.flyte.org/flyteorg/flytecopilot:v0.0.9
  

default-input-path string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Default path where the volume should be mounted

**Default Value**: 

.. code-block:: yaml

  /var/flyte/inputs
  

default-output-path string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Default path where the volume should be mounted

**Default Value**: 

.. code-block:: yaml

  /var/flyte/outputs
  

input-vol-name string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Name of the data volume that is created for storing inputs

**Default Value**: 

.. code-block:: yaml

  flyte-inputs
  

output-vol-name string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Name of the data volume that is created for storing outputs

**Default Value**: 

.. code-block:: yaml

  flyte-outputs
  

start-timeout `config.Duration`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  1m0s
  

cpu string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Used to set cpu for co-pilot containers

**Default Value**: 

.. code-block:: yaml

  500m
  

memory string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Used to set memory for co-pilot containers

**Default Value**: 

.. code-block:: yaml

  128Mi
  

storage string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Default storage limit for individual inputs / outputs

**Default Value**: 

.. code-block:: yaml

  ""
  

resource.Quantity
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

i `resource.int64Amount`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  {}
  

d `resource.infDecAmount`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  <nil>
  

s string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "1"
  

Format string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  DecimalSI
  

resource.infDecAmount
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Dec inf.Dec
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

resource.int64Amount
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

value int64
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "1"
  

scale int32
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

k8s.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

scheduler string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Decides the scheduler to use when launching array-pods.

**Default Value**: 

.. code-block:: yaml

  ""
  

maxErrorLength int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Determines the maximum length of the error string returned for the array.

**Default Value**: 

.. code-block:: yaml

  "1000"
  

maxArrayJobSize int64
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Maximum size of array job.

**Default Value**: 

.. code-block:: yaml

  "5000"
  

resourceConfig `k8s.ResourceConfig`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  limit: 0
  primaryLabel: ""
  

remoteClusterConfig `k8s.ClusterConfig`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  auth:
    certPath: ""
    tokenPath: ""
    type: ""
  enabled: false
  endpoint: ""
  name: ""
  

node-selector map[string]string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

tolerations []v1.Toleration
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

namespaceTemplate string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

OutputAssembler `workqueue.Config`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  maxItems: 100000
  maxRetries: 5
  workers: 10
  

ErrorAssembler `workqueue.Config`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  maxItems: 100000
  maxRetries: 5
  workers: 10
  

logs `k8s.LogConfig`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Config for log links for k8s array jobs.

**Default Value**: 

.. code-block:: yaml

  config:
    cloudwatch-enabled: false
    cloudwatch-log-group: ""
    cloudwatch-region: ""
    cloudwatch-template-uri: ""
    gcp-project: ""
    kubernetes-enabled: false
    kubernetes-template-uri: ""
    kubernetes-url: ""
    stackdriver-enabled: false
    stackdriver-logresourcename: ""
    stackdriver-template-uri: ""
    templates: null
  

k8s.ClusterConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

name string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Friendly name of the remote cluster

**Default Value**: 

.. code-block:: yaml

  ""
  

endpoint string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Remote K8s cluster endpoint

**Default Value**: 

.. code-block:: yaml

  ""
  

auth `k8s.Auth`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  certPath: ""
  tokenPath: ""
  type: ""
  

enabled bool
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Boolean flag to enable or disable

**Default Value**: 

.. code-block:: yaml

  "false"
  

k8s.Auth
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

type string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Authentication type

**Default Value**: 

.. code-block:: yaml

  ""
  

tokenPath string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Token path

**Default Value**: 

.. code-block:: yaml

  ""
  

certPath string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Certificate path

**Default Value**: 

.. code-block:: yaml

  ""
  

k8s.LogConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

config `logs.LogConfig (config)`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines the log config for k8s logs.

**Default Value**: 

.. code-block:: yaml

  cloudwatch-enabled: false
  cloudwatch-log-group: ""
  cloudwatch-region: ""
  cloudwatch-template-uri: ""
  gcp-project: ""
  kubernetes-enabled: false
  kubernetes-template-uri: ""
  kubernetes-url: ""
  stackdriver-enabled: false
  stackdriver-logresourcename: ""
  stackdriver-template-uri: ""
  templates: null
  

logs.LogConfig (config)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

cloudwatch-enabled bool
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Enable Cloudwatch Logging

**Default Value**: 

.. code-block:: yaml

  "false"
  

cloudwatch-region string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: AWS region in which Cloudwatch logs are stored.

**Default Value**: 

.. code-block:: yaml

  ""
  

cloudwatch-log-group string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Log group to which streams are associated.

**Default Value**: 

.. code-block:: yaml

  ""
  

cloudwatch-template-uri string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Template Uri to use when building cloudwatch log links

**Default Value**: 

.. code-block:: yaml

  ""
  

kubernetes-enabled bool
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Enable Kubernetes Logging

**Default Value**: 

.. code-block:: yaml

  "false"
  

kubernetes-url string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Console URL for Kubernetes logs

**Default Value**: 

.. code-block:: yaml

  ""
  

kubernetes-template-uri string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Template Uri to use when building kubernetes log links

**Default Value**: 

.. code-block:: yaml

  ""
  

stackdriver-enabled bool
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Enable Log-links to stackdriver

**Default Value**: 

.. code-block:: yaml

  "false"
  

gcp-project string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Name of the project in GCP

**Default Value**: 

.. code-block:: yaml

  ""
  

stackdriver-logresourcename string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Name of the logresource in stackdriver

**Default Value**: 

.. code-block:: yaml

  ""
  

stackdriver-template-uri string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Template Uri to use when building stackdriver log links

**Default Value**: 

.. code-block:: yaml

  ""
  

templates []logs.TemplateLogPluginConfig
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

k8s.ResourceConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

primaryLabel string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: PrimaryLabel of a given service cluster

**Default Value**: 

.. code-block:: yaml

  ""
  

limit int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Resource quota (in the number of outstanding requests) for the cluster

**Default Value**: 

.. code-block:: yaml

  "0"
  

logs.LogConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

cloudwatch-enabled bool
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Enable Cloudwatch Logging

**Default Value**: 

.. code-block:: yaml

  "false"
  

cloudwatch-region string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: AWS region in which Cloudwatch logs are stored.

**Default Value**: 

.. code-block:: yaml

  ""
  

cloudwatch-log-group string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Log group to which streams are associated.

**Default Value**: 

.. code-block:: yaml

  ""
  

cloudwatch-template-uri string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Template Uri to use when building cloudwatch log links

**Default Value**: 

.. code-block:: yaml

  ""
  

kubernetes-enabled bool
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Enable Kubernetes Logging

**Default Value**: 

.. code-block:: yaml

  "false"
  

kubernetes-url string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Console URL for Kubernetes logs

**Default Value**: 

.. code-block:: yaml

  ""
  

kubernetes-template-uri string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Template Uri to use when building kubernetes log links

**Default Value**: 

.. code-block:: yaml

  ""
  

stackdriver-enabled bool
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Enable Log-links to stackdriver

**Default Value**: 

.. code-block:: yaml

  "false"
  

gcp-project string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Name of the project in GCP

**Default Value**: 

.. code-block:: yaml

  ""
  

stackdriver-logresourcename string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Name of the logresource in stackdriver

**Default Value**: 

.. code-block:: yaml

  ""
  

stackdriver-template-uri string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Template Uri to use when building stackdriver log links

**Default Value**: 

.. code-block:: yaml

  ""
  

templates []logs.TemplateLogPluginConfig
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

snowflake.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

webApi `webapi.PluginConfig`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines config for the base WebAPI plugin.

**Default Value**: 

.. code-block:: yaml

  caching:
    maxSystemFailures: 5
    resyncInterval: 30s
    size: 500000
    workers: 10
  readRateLimiter:
    burst: 100
    qps: 10
  resourceMeta: null
  resourceQuotas:
    default: 1000
  writeRateLimiter:
    burst: 100
    qps: 10
  

resourceConstraints `core.ResourceConstraintsSpec`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  NamespaceScopeResourceConstraint:
    Value: 50
  ProjectScopeResourceConstraint:
    Value: 100
  

defaultWarehouse string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines the default warehouse to use when running on Snowflake unless overwritten by the task.

**Default Value**: 

.. code-block:: yaml

  COMPUTE_WH
  

snowflakeTokenKey string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Name of the key where to find Snowflake token in the secret manager.

**Default Value**: 

.. code-block:: yaml

  FLYTE_SNOWFLAKE_CLIENT_TOKEN
  

snowflakeEndpoint string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

spark.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

spark-config-default map[string]string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

spark-history-server-url string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: URL for SparkHistory Server that each job will publish the execution history to.

**Default Value**: 

.. code-block:: yaml

  ""
  

features []spark.Feature
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

logs `spark.LogConfig`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Config for log links for spark applications.

**Default Value**: 

.. code-block:: yaml

  all-user:
    cloudwatch-enabled: false
    cloudwatch-log-group: ""
    cloudwatch-region: ""
    cloudwatch-template-uri: ""
    gcp-project: ""
    kubernetes-enabled: false
    kubernetes-template-uri: ""
    kubernetes-url: ""
    stackdriver-enabled: false
    stackdriver-logresourcename: ""
    stackdriver-template-uri: ""
    templates: null
  mixed:
    cloudwatch-enabled: false
    cloudwatch-log-group: ""
    cloudwatch-region: ""
    cloudwatch-template-uri: ""
    gcp-project: ""
    kubernetes-enabled: true
    kubernetes-template-uri: http://localhost:30082/#!/log/{{ .namespace }}/{{ .podName
      }}/pod?namespace={{ .namespace }}
    kubernetes-url: ""
    stackdriver-enabled: false
    stackdriver-logresourcename: ""
    stackdriver-template-uri: ""
    templates: null
  system:
    cloudwatch-enabled: false
    cloudwatch-log-group: ""
    cloudwatch-region: ""
    cloudwatch-template-uri: ""
    gcp-project: ""
    kubernetes-enabled: false
    kubernetes-template-uri: ""
    kubernetes-url: ""
    stackdriver-enabled: false
    stackdriver-logresourcename: ""
    stackdriver-template-uri: ""
    templates: null
  user:
    cloudwatch-enabled: false
    cloudwatch-log-group: ""
    cloudwatch-region: ""
    cloudwatch-template-uri: ""
    gcp-project: ""
    kubernetes-enabled: false
    kubernetes-template-uri: ""
    kubernetes-url: ""
    stackdriver-enabled: false
    stackdriver-logresourcename: ""
    stackdriver-template-uri: ""
    templates: null
  

spark.LogConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

mixed `logs.LogConfig`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines the log config that's not split into user/system.

**Default Value**: 

.. code-block:: yaml

  cloudwatch-enabled: false
  cloudwatch-log-group: ""
  cloudwatch-region: ""
  cloudwatch-template-uri: ""
  gcp-project: ""
  kubernetes-enabled: true
  kubernetes-template-uri: http://localhost:30082/#!/log/{{ .namespace }}/{{ .podName
    }}/pod?namespace={{ .namespace }}
  kubernetes-url: ""
  stackdriver-enabled: false
  stackdriver-logresourcename: ""
  stackdriver-template-uri: ""
  templates: null
  

user `logs.LogConfig`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines the log config for user logs.

**Default Value**: 

.. code-block:: yaml

  cloudwatch-enabled: false
  cloudwatch-log-group: ""
  cloudwatch-region: ""
  cloudwatch-template-uri: ""
  gcp-project: ""
  kubernetes-enabled: false
  kubernetes-template-uri: ""
  kubernetes-url: ""
  stackdriver-enabled: false
  stackdriver-logresourcename: ""
  stackdriver-template-uri: ""
  templates: null
  

system `logs.LogConfig`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Defines the log config for system logs.

**Default Value**: 

.. code-block:: yaml

  cloudwatch-enabled: false
  cloudwatch-log-group: ""
  cloudwatch-region: ""
  cloudwatch-template-uri: ""
  gcp-project: ""
  kubernetes-enabled: false
  kubernetes-template-uri: ""
  kubernetes-url: ""
  stackdriver-enabled: false
  stackdriver-logresourcename: ""
  stackdriver-template-uri: ""
  templates: null
  

all-user `logs.LogConfig`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: All user logs across driver and executors.

**Default Value**: 

.. code-block:: yaml

  cloudwatch-enabled: false
  cloudwatch-log-group: ""
  cloudwatch-region: ""
  cloudwatch-template-uri: ""
  gcp-project: ""
  kubernetes-enabled: false
  kubernetes-template-uri: ""
  kubernetes-url: ""
  stackdriver-enabled: false
  stackdriver-logresourcename: ""
  stackdriver-template-uri: ""
  templates: null
  

Section: propeller
================================================================================

kube-config string
--------------------------------------------------------------------------------

**Description**: Path to kubernetes client config file.

**Default Value**: 

.. code-block:: yaml

  ""
  

master string
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

workers int
--------------------------------------------------------------------------------

**Description**: Number of threads to process workflows

**Default Value**: 

.. code-block:: yaml

  "20"
  

workflow-reeval-duration `config.Duration`_
--------------------------------------------------------------------------------

**Description**: Frequency of re-evaluating workflows

**Default Value**: 

.. code-block:: yaml

  10s
  

downstream-eval-duration `config.Duration`_
--------------------------------------------------------------------------------

**Description**: Frequency of re-evaluating downstream tasks

**Default Value**: 

.. code-block:: yaml

  30s
  

limit-namespace string
--------------------------------------------------------------------------------

**Description**: Namespaces to watch for this propeller

**Default Value**: 

.. code-block:: yaml

  all
  

prof-port `config.Port`_
--------------------------------------------------------------------------------

**Description**: Profiler port

**Default Value**: 

.. code-block:: yaml

  10254
  

metadata-prefix string
--------------------------------------------------------------------------------

**Description**: MetadataPrefix should be used if all the metadata for Flyte executions should be stored under a specific prefix in CloudStorage. If not specified, the data will be stored in the base container directly.

**Default Value**: 

.. code-block:: yaml

  metadata/propeller
  

rawoutput-prefix string
--------------------------------------------------------------------------------

**Description**: a fully qualified storage path of the form s3://flyte/abc/..., where all data sandboxes should be stored.

**Default Value**: 

.. code-block:: yaml

  ""
  

queue `config.CompositeQueueConfig`_
--------------------------------------------------------------------------------

**Description**: Workflow workqueue configuration, affects the way the work is consumed from the queue.

**Default Value**: 

.. code-block:: yaml

  batch-size: -1
  batching-interval: 1s
  queue:
    base-delay: 5s
    capacity: 1000
    max-delay: 1m0s
    rate: 100
    type: maxof
  sub-queue:
    base-delay: 0s
    capacity: 1000
    max-delay: 0s
    rate: 100
    type: bucket
  type: batch
  

metrics-prefix string
--------------------------------------------------------------------------------

**Description**: An optional prefix for all published metrics.

**Default Value**: 

.. code-block:: yaml

  flyte
  

enable-admin-launcher bool
--------------------------------------------------------------------------------

**Description**: Enable remote Workflow launcher to Admin

**Default Value**: 

.. code-block:: yaml

  "true"
  

max-workflow-retries int
--------------------------------------------------------------------------------

**Description**: Maximum number of retries per workflow

**Default Value**: 

.. code-block:: yaml

  "10"
  

max-ttl-hours int
--------------------------------------------------------------------------------

**Description**: Maximum number of hours a completed workflow should be retained. Number between 1-23 hours

**Default Value**: 

.. code-block:: yaml

  "23"
  

gc-interval `config.Duration`_
--------------------------------------------------------------------------------

**Description**: Run periodic GC every 30 minutes

**Default Value**: 

.. code-block:: yaml

  30m0s
  

leader-election `config.LeaderElectionConfig`_
--------------------------------------------------------------------------------

**Description**: Config for leader election.

**Default Value**: 

.. code-block:: yaml

  enabled: false
  lease-duration: 15s
  lock-config-map:
    Name: ""
    Namespace: ""
  renew-deadline: 10s
  retry-period: 2s
  

publish-k8s-events bool
--------------------------------------------------------------------------------

**Description**: Enable events publishing to K8s events API.

**Default Value**: 

.. code-block:: yaml

  "false"
  

max-output-size-bytes int64
--------------------------------------------------------------------------------

**Description**: Maximum size of outputs per task

**Default Value**: 

.. code-block:: yaml

  "10485760"
  

kube-client-config `config.KubeClientConfig`_
--------------------------------------------------------------------------------

**Description**: Configuration to control the Kubernetes client

**Default Value**: 

.. code-block:: yaml

  burst: 25
  qps: 100
  timeout: 30s
  

node-config `config.NodeConfig`_
--------------------------------------------------------------------------------

**Description**: config for a workflow node

**Default Value**: 

.. code-block:: yaml

  default-deadlines:
    node-active-deadline: 48h0m0s
    node-execution-deadline: 48h0m0s
    workflow-active-deadline: 72h0m0s
  interruptible-failure-threshold: 1
  max-node-retries-system-failures: 3
  

max-streak-length int
--------------------------------------------------------------------------------

**Description**: Maximum number of consecutive rounds that one propeller worker can use for one workflow - >1 => turbo-mode is enabled.

**Default Value**: 

.. code-block:: yaml

  "8"
  

event-config `config.EventConfig`_
--------------------------------------------------------------------------------

**Description**: Configures execution event behavior.

**Default Value**: 

.. code-block:: yaml

  fallback-to-output-reference: false
  raw-output-policy: reference
  

admin-launcher `launchplan.AdminConfig`_
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  burst: 10
  cacheSize: 10000
  tps: 100
  workers: 10
  

resourcemanager `config.Config (resourcemanager)`_
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  redis:
    hostKey: ""
    hostPath: ""
    hostPaths: []
    maxRetries: 0
    primaryName: ""
  resourceMaxQuota: 1000
  type: noop
  

workflowstore `workflowstore.Config`_
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  policy: ResourceVersionCache
  

config.CompositeQueueConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

type string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Type of composite queue to use for the WorkQueue

**Default Value**: 

.. code-block:: yaml

  batch
  

queue `config.WorkqueueConfig`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Workflow workqueue configuration, affects the way the work is consumed from the queue.

**Default Value**: 

.. code-block:: yaml

  base-delay: 5s
  capacity: 1000
  max-delay: 1m0s
  rate: 100
  type: maxof
  

sub-queue `config.WorkqueueConfig`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: SubQueue configuration, affects the way the nodes cause the top-level Work to be re-evaluated.

**Default Value**: 

.. code-block:: yaml

  base-delay: 0s
  capacity: 1000
  max-delay: 0s
  rate: 100
  type: bucket
  

batching-interval `config.Duration`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Duration for which downstream updates are buffered

**Default Value**: 

.. code-block:: yaml

  1s
  

batch-size int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "-1"
  

config.WorkqueueConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

type string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Type of RateLimiter to use for the WorkQueue

**Default Value**: 

.. code-block:: yaml

  maxof
  

base-delay `config.Duration`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: base backoff delay for failure

**Default Value**: 

.. code-block:: yaml

  5s
  

max-delay `config.Duration`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Max backoff delay for failure

**Default Value**: 

.. code-block:: yaml

  1m0s
  

rate int64
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Bucket Refill rate per second

**Default Value**: 

.. code-block:: yaml

  "100"
  

capacity int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Bucket capacity as number of items

**Default Value**: 

.. code-block:: yaml

  "1000"
  

config.Config (resourcemanager)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

type string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Which resource manager to use

**Default Value**: 

.. code-block:: yaml

  noop
  

resourceMaxQuota int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Global limit for concurrent Qubole queries

**Default Value**: 

.. code-block:: yaml

  "1000"
  

redis `config.RedisConfig`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Config for Redis resourcemanager.

**Default Value**: 

.. code-block:: yaml

  hostKey: ""
  hostPath: ""
  hostPaths: []
  maxRetries: 0
  primaryName: ""
  

config.RedisConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

hostPaths []string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Redis hosts locations.

**Default Value**: 

.. code-block:: yaml

  []
  

primaryName string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Redis primary name, fill in only if you are connecting to a redis sentinel cluster.

**Default Value**: 

.. code-block:: yaml

  ""
  

hostPath string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Redis host location

**Default Value**: 

.. code-block:: yaml

  ""
  

hostKey string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Key for local Redis access

**Default Value**: 

.. code-block:: yaml

  ""
  

maxRetries int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: See Redis client options for more info

**Default Value**: 

.. code-block:: yaml

  "0"
  

config.EventConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

raw-output-policy string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: How output data should be passed along in execution events.

**Default Value**: 

.. code-block:: yaml

  reference
  

fallback-to-output-reference bool
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Whether output data should be sent by reference when it is too large to be sent inline in execution events.

**Default Value**: 

.. code-block:: yaml

  "false"
  

config.KubeClientConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

qps float32
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "100"
  

burst int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Max burst rate for throttle. 0 defaults to 10

**Default Value**: 

.. code-block:: yaml

  "25"
  

timeout `config.Duration`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Max duration allowed for every request to KubeAPI before giving up. 0 implies no timeout.

**Default Value**: 

.. code-block:: yaml

  30s
  

config.LeaderElectionConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

enabled bool
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Enables/Disables leader election.

**Default Value**: 

.. code-block:: yaml

  "false"
  

lock-config-map `types.NamespacedName`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: ConfigMap namespace/name to use for resource lock.

**Default Value**: 

.. code-block:: yaml

  Name: ""
  Namespace: ""
  

lease-duration `config.Duration`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Duration that non-leader candidates will wait to force acquire leadership. This is measured against time of last observed ack.

**Default Value**: 

.. code-block:: yaml

  15s
  

renew-deadline `config.Duration`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Duration that the acting master will retry refreshing leadership before giving up.

**Default Value**: 

.. code-block:: yaml

  10s
  

retry-period `config.Duration`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Duration the LeaderElector clients should wait between tries of actions.

**Default Value**: 

.. code-block:: yaml

  2s
  

types.NamespacedName
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Namespace string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Name string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

config.NodeConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

default-deadlines `config.DefaultDeadlines`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Default value for timeouts

**Default Value**: 

.. code-block:: yaml

  node-active-deadline: 48h0m0s
  node-execution-deadline: 48h0m0s
  workflow-active-deadline: 72h0m0s
  

max-node-retries-system-failures int64
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Maximum number of retries per node for node failure due to infra issues

**Default Value**: 

.. code-block:: yaml

  "3"
  

interruptible-failure-threshold int64
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: number of failures for a node to be still considered interruptible'

**Default Value**: 

.. code-block:: yaml

  "1"
  

config.DefaultDeadlines
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

node-execution-deadline `config.Duration`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Default value of node execution timeout

**Default Value**: 

.. code-block:: yaml

  48h0m0s
  

node-active-deadline `config.Duration`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Default value of node timeout

**Default Value**: 

.. code-block:: yaml

  48h0m0s
  

workflow-active-deadline `config.Duration`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Default value of workflow timeout

**Default Value**: 

.. code-block:: yaml

  72h0m0s
  

config.Port
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

port int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "10254"
  

launchplan.AdminConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

tps int64
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: The maximum number of transactions per second to flyte admin from this client.

**Default Value**: 

.. code-block:: yaml

  "100"
  

burst int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Maximum burst for throttle

**Default Value**: 

.. code-block:: yaml

  "10"
  

cacheSize int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Maximum cache in terms of number of items stored.

**Default Value**: 

.. code-block:: yaml

  "10000"
  

workers int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Number of parallel workers to work on the queue.

**Default Value**: 

.. code-block:: yaml

  "10"
  

workflowstore.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

policy string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Workflow Store Policy to initialize

**Default Value**: 

.. code-block:: yaml

  ResourceVersionCache
  

Section: secrets
================================================================================

secrets-prefix string
--------------------------------------------------------------------------------

**Description**: Prefix where to look for secrets file

**Default Value**: 

.. code-block:: yaml

  /etc/secrets
  

env-prefix string
--------------------------------------------------------------------------------

**Description**: Prefix for environment variables

**Default Value**: 

.. code-block:: yaml

  FLYTE_SECRET_
  

Section: storage
================================================================================

type string
--------------------------------------------------------------------------------

**Description**: Sets the type of storage to configure [s3/minio/local/mem/stow].

**Default Value**: 

.. code-block:: yaml

  s3
  

connection `storage.ConnectionConfig`_
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  access-key: ""
  auth-type: iam
  disable-ssl: false
  endpoint: ""
  region: us-east-1
  secret-key: ""
  

stow `storage.StowConfig`_
--------------------------------------------------------------------------------

**Description**: Storage config for stow backend.

**Default Value**: 

.. code-block:: yaml

  {}
  

container string
--------------------------------------------------------------------------------

**Description**: Initial container (in s3 a bucket) to create -if it doesn't exist-.'

**Default Value**: 

.. code-block:: yaml

  ""
  

enable-multicontainer bool
--------------------------------------------------------------------------------

**Description**: If this is true, then the container argument is overlooked and redundant. This config will automatically open new connections to new containers/buckets as they are encountered

**Default Value**: 

.. code-block:: yaml

  "false"
  

cache `storage.CachingConfig`_
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  max_size_mbs: 0
  target_gc_percent: 0
  

limits `storage.LimitsConfig`_
--------------------------------------------------------------------------------

**Description**: Sets limits for stores.

**Default Value**: 

.. code-block:: yaml

  maxDownloadMBs: 2
  

defaultHttpClient `storage.HTTPClientConfig`_
--------------------------------------------------------------------------------

**Description**: Sets the default http client config.

**Default Value**: 

.. code-block:: yaml

  headers: null
  timeout: 0s
  

storage.CachingConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

max_size_mbs int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Maximum size of the cache where the Blob store data is cached in-memory. If not specified or set to 0, cache is not used

**Default Value**: 

.. code-block:: yaml

  "0"
  

target_gc_percent int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Sets the garbage collection target percentage.

**Default Value**: 

.. code-block:: yaml

  "0"
  

storage.ConnectionConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

endpoint `config.URL`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: URL for storage client to connect to.

**Default Value**: 

.. code-block:: yaml

  ""
  

auth-type string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Auth Type to use [iam,accesskey].

**Default Value**: 

.. code-block:: yaml

  iam
  

access-key string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Access key to use. Only required when authtype is set to accesskey.

**Default Value**: 

.. code-block:: yaml

  ""
  

secret-key string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Secret to use when accesskey is set.

**Default Value**: 

.. code-block:: yaml

  ""
  

region string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Region to connect to.

**Default Value**: 

.. code-block:: yaml

  us-east-1
  

disable-ssl bool
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Disables SSL connection. Should only be used for development.

**Default Value**: 

.. code-block:: yaml

  "false"
  

storage.HTTPClientConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

headers map[string][]string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

timeout `config.Duration`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Sets time out on the http client.

**Default Value**: 

.. code-block:: yaml

  0s
  

storage.LimitsConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

maxDownloadMBs int64
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Maximum allowed download size (in MBs) per call.

**Default Value**: 

.. code-block:: yaml

  "2"
  

storage.StowConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

kind string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Kind of Stow backend to use. Refer to github/graymeta/stow

**Default Value**: 

.. code-block:: yaml

  ""
  

config map[string]string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Configuration for stow backend. Refer to github/graymeta/stow

**Default Value**: 

.. code-block:: yaml

  {}
  

Section: tasks
================================================================================

task-plugins `config.TaskPluginConfig`_
--------------------------------------------------------------------------------

**Description**: Task plugin configuration

**Default Value**: 

.. code-block:: yaml

  default-for-task-types: {}
  enabled-plugins: []
  

max-plugin-phase-versions int32
--------------------------------------------------------------------------------

**Description**: Maximum number of plugin phase versions allowed for one phase.

**Default Value**: 

.. code-block:: yaml

  "100000"
  

barrier `config.BarrierConfig`_
--------------------------------------------------------------------------------

**Description**: Config for Barrier implementation

**Default Value**: 

.. code-block:: yaml

  cache-size: 10000
  cache-ttl: 30m0s
  enabled: true
  

backoff `config.BackOffConfig`_
--------------------------------------------------------------------------------

**Description**: Config for Exponential BackOff implementation

**Default Value**: 

.. code-block:: yaml

  base-second: 2
  max-duration: 10m0s
  

maxLogMessageLength int
--------------------------------------------------------------------------------

**Description**: Max length of error message.

**Default Value**: 

.. code-block:: yaml

  "2048"
  

config.BackOffConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

base-second int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: The number of seconds representing the base duration of the exponential backoff

**Default Value**: 

.. code-block:: yaml

  "2"
  

max-duration `config.Duration`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: The cap of the backoff duration

**Default Value**: 

.. code-block:: yaml

  10m0s
  

config.BarrierConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

enabled bool
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Enable Barrier transitions using inmemory context

**Default Value**: 

.. code-block:: yaml

  "true"
  

cache-size int
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Max number of barrier to preserve in memory

**Default Value**: 

.. code-block:: yaml

  "10000"
  

cache-ttl `config.Duration`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Max duration that a barrier would be respected if the process is not restarted. This should account for time required to store the record into persistent storage (across multiple rounds.

**Default Value**: 

.. code-block:: yaml

  30m0s
  

config.TaskPluginConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

enabled-plugins []string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: deprecated

**Default Value**: 

.. code-block:: yaml

  []
  

default-for-task-types map[string]string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  {}
  

Section: webhook
================================================================================

metrics-prefix string
--------------------------------------------------------------------------------

**Description**: An optional prefix for all published metrics.

**Default Value**: 

.. code-block:: yaml

  'flyte:'
  

certDir string
--------------------------------------------------------------------------------

**Description**: Certificate directory to use to write generated certs. Defaults to /etc/webhook/certs/

**Default Value**: 

.. code-block:: yaml

  /etc/webhook/certs
  

listenPort int
--------------------------------------------------------------------------------

**Description**: The port to use to listen to webhook calls. Defaults to 9443

**Default Value**: 

.. code-block:: yaml

  "9443"
  

serviceName string
--------------------------------------------------------------------------------

**Description**: The name of the webhook service.

**Default Value**: 

.. code-block:: yaml

  flyte-pod-webhook
  

secretName string
--------------------------------------------------------------------------------

**Description**: Secret name to write generated certs to.

**Default Value**: 

.. code-block:: yaml

  flyte-pod-webhook
  

secretManagerType int
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  K8s
  

awsSecretManager `config.AWSSecretManagerConfig`_
--------------------------------------------------------------------------------

**Description**: AWS Secret Manager config.

**Default Value**: 

.. code-block:: yaml

  resources:
    limits:
      cpu: 200m
      memory: 500Mi
    requests:
      cpu: 200m
      memory: 500Mi
  sidecarImage: docker.io/amazon/aws-secrets-manager-secret-sidecar:v0.1.4
  

config.AWSSecretManagerConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

sidecarImage string
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Description**: Specifies the sidecar docker image to use

**Default Value**: 

.. code-block:: yaml

  docker.io/amazon/aws-secrets-manager-secret-sidecar:v0.1.4
  

resources `v1.ResourceRequirements`_
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  limits:
    cpu: 200m
    memory: 500Mi
  requests:
    cpu: 200m
    memory: 500Mi
  

v1.ResourceRequirements
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

limits v1.ResourceList
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  cpu: 200m
  memory: 500Mi
  

requests v1.ResourceList
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  cpu: 200m
  memory: 500Mi
  

