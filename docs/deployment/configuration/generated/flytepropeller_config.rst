.. _flytepropeller-config-specification:

#########################################
Flyte Propeller Configuration
#########################################

- `admin <#section-admin>`_

- `catalog-cache <#section-catalog-cache>`_

- `event <#section-event>`_

- `logger <#section-logger>`_

- `otel <#section-otel>`_

- `plugins <#section-plugins>`_

- `propeller <#section-propeller>`_

- `secrets <#section-secrets>`_

- `storage <#section-storage>`_

- `tasks <#section-tasks>`_

- `webhook <#section-webhook>`_

Section: admin
========================================================================================================================

endpoint (`config.URL`_)
------------------------------------------------------------------------------------------------------------------------

For admin types, specify where the uri of the service is located.

**Default Value**: 

.. code-block:: yaml

  ""
  

insecure (bool)
------------------------------------------------------------------------------------------------------------------------

Use insecure connection.

**Default Value**: 

.. code-block:: yaml

  "false"
  

insecureSkipVerify (bool)
------------------------------------------------------------------------------------------------------------------------

InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name. Caution : shouldn't be use for production usecases'

**Default Value**: 

.. code-block:: yaml

  "false"
  

caCertFilePath (string)
------------------------------------------------------------------------------------------------------------------------

Use specified certificate file to verify the admin server peer.

**Default Value**: 

.. code-block:: yaml

  ""
  

maxBackoffDelay (`config.Duration`_)
------------------------------------------------------------------------------------------------------------------------

Max delay for grpc backoff

**Default Value**: 

.. code-block:: yaml

  8s
  

perRetryTimeout (`config.Duration`_)
------------------------------------------------------------------------------------------------------------------------

gRPC per retry timeout

**Default Value**: 

.. code-block:: yaml

  15s
  

maxRetries (int)
------------------------------------------------------------------------------------------------------------------------

Max number of gRPC retries

**Default Value**: 

.. code-block:: yaml

  "4"
  

authType (uint8)
------------------------------------------------------------------------------------------------------------------------

Type of OAuth2 flow used for communicating with admin.ClientSecret,Pkce,ExternalCommand are valid values

**Default Value**: 

.. code-block:: yaml

  ClientSecret
  

tokenRefreshWindow (`config.Duration`_)
------------------------------------------------------------------------------------------------------------------------

Max duration between token refresh attempt and token expiry.

**Default Value**: 

.. code-block:: yaml

  0s
  

useAuth (bool)
------------------------------------------------------------------------------------------------------------------------

Deprecated: Auth will be enabled/disabled based on admin's dynamically discovered information.

**Default Value**: 

.. code-block:: yaml

  "false"
  

clientId (string)
------------------------------------------------------------------------------------------------------------------------

Client ID

**Default Value**: 

.. code-block:: yaml

  flytepropeller
  

clientSecretLocation (string)
------------------------------------------------------------------------------------------------------------------------

File containing the client secret

**Default Value**: 

.. code-block:: yaml

  /etc/secrets/client_secret
  

clientSecretEnvVar (string)
------------------------------------------------------------------------------------------------------------------------

Environment variable containing the client secret

**Default Value**: 

.. code-block:: yaml

  ""
  

scopes ([]string)
------------------------------------------------------------------------------------------------------------------------

List of scopes to request

**Default Value**: 

.. code-block:: yaml

  []
  

useAudienceFromAdmin (bool)
------------------------------------------------------------------------------------------------------------------------

Use Audience configured from admins public endpoint config.

**Default Value**: 

.. code-block:: yaml

  "false"
  

audience (string)
------------------------------------------------------------------------------------------------------------------------

Audience to use when initiating OAuth2 authorization requests.

**Default Value**: 

.. code-block:: yaml

  ""
  

authorizationServerUrl (string)
------------------------------------------------------------------------------------------------------------------------

This is the URL to your IdP's authorization server. It'll default to Endpoint

**Default Value**: 

.. code-block:: yaml

  ""
  

tokenUrl (string)
------------------------------------------------------------------------------------------------------------------------

OPTIONAL: Your IdP's token endpoint. It'll be discovered from flyte admin's OAuth Metadata endpoint if not provided.

**Default Value**: 

.. code-block:: yaml

  ""
  

authorizationHeader (string)
------------------------------------------------------------------------------------------------------------------------

Custom metadata header to pass JWT

**Default Value**: 

.. code-block:: yaml

  ""
  

pkceConfig (`pkce.Config`_)
------------------------------------------------------------------------------------------------------------------------

Config for Pkce authentication flow.

**Default Value**: 

.. code-block:: yaml

  refreshTime: 5m0s
  timeout: 2m0s
  

deviceFlowConfig (`deviceflow.Config`_)
------------------------------------------------------------------------------------------------------------------------

Config for Device authentication flow.

**Default Value**: 

.. code-block:: yaml

  pollInterval: 5s
  refreshTime: 5m0s
  timeout: 10m0s
  

command ([]string)
------------------------------------------------------------------------------------------------------------------------

Command for external authentication token generation

**Default Value**: 

.. code-block:: yaml

  []
  

proxyCommand ([]string)
------------------------------------------------------------------------------------------------------------------------

Command for external proxy-authorization token generation

**Default Value**: 

.. code-block:: yaml

  []
  

defaultServiceConfig (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

httpProxyURL (`config.URL`_)
------------------------------------------------------------------------------------------------------------------------

OPTIONAL: HTTP Proxy to be used for OAuth requests.

**Default Value**: 

.. code-block:: yaml

  ""
  

config.Duration
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Duration (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  8s
  

config.URL
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

URL (`url.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ForceQuery: false
  Fragment: ""
  Host: ""
  OmitHost: false
  Opaque: ""
  Path: ""
  RawFragment: ""
  RawPath: ""
  RawQuery: ""
  Scheme: ""
  User: null
  

url.URL
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Scheme (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Opaque (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

User (url.Userinfo)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

Host (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Path (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

RawPath (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

OmitHost (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "false"
  

ForceQuery (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "false"
  

RawQuery (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Fragment (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

RawFragment (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

deviceflow.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

refreshTime (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

grace period from the token expiry after which it would refresh the token.

**Default Value**: 

.. code-block:: yaml

  5m0s
  

timeout (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

amount of time the device flow should complete or else it will be cancelled.

**Default Value**: 

.. code-block:: yaml

  10m0s
  

pollInterval (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

amount of time the device flow would poll the token endpoint if auth server doesn't return a polling interval. Okta and google IDP do return an interval'

**Default Value**: 

.. code-block:: yaml

  5s
  

pkce.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

timeout (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Amount of time the browser session would be active for authentication from client app.

**Default Value**: 

.. code-block:: yaml

  2m0s
  

refreshTime (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

grace period from the token expiry after which it would refresh the token.

**Default Value**: 

.. code-block:: yaml

  5m0s
  

Section: catalog-cache
========================================================================================================================

type (string)
------------------------------------------------------------------------------------------------------------------------

Catalog Implementation to use

**Default Value**: 

.. code-block:: yaml

  noop
  

endpoint (string)
------------------------------------------------------------------------------------------------------------------------

Endpoint for catalog service

**Default Value**: 

.. code-block:: yaml

  ""
  

insecure (bool)
------------------------------------------------------------------------------------------------------------------------

Use insecure grpc connection

**Default Value**: 

.. code-block:: yaml

  "false"
  

max-cache-age (`config.Duration`_)
------------------------------------------------------------------------------------------------------------------------

Cache entries past this age will incur cache miss. 0 means cache never expires

**Default Value**: 

.. code-block:: yaml

  0s
  

use-admin-auth (bool)
------------------------------------------------------------------------------------------------------------------------

Use the same gRPC credentials option as the flyteadmin client

**Default Value**: 

.. code-block:: yaml

  "false"
  

default-service-config (string)
------------------------------------------------------------------------------------------------------------------------

Set the default service config for the catalog gRPC client

**Default Value**: 

.. code-block:: yaml

  ""
  

Section: event
========================================================================================================================

type (string)
------------------------------------------------------------------------------------------------------------------------

Sets the type of EventSink to configure [log/admin/file].

**Default Value**: 

.. code-block:: yaml

  admin
  

file-path (string)
------------------------------------------------------------------------------------------------------------------------

For file types, specify where the file should be located.

**Default Value**: 

.. code-block:: yaml

  ""
  

rate (int64)
------------------------------------------------------------------------------------------------------------------------

Max rate at which events can be recorded per second.

**Default Value**: 

.. code-block:: yaml

  "500"
  

capacity (int)
------------------------------------------------------------------------------------------------------------------------

The max bucket size for event recording tokens.

**Default Value**: 

.. code-block:: yaml

  "1000"
  

Section: logger
========================================================================================================================

show-source (bool)
------------------------------------------------------------------------------------------------------------------------

Includes source code location in logs.

**Default Value**: 

.. code-block:: yaml

  "false"
  

mute (bool)
------------------------------------------------------------------------------------------------------------------------

Mutes all logs regardless of severity. Intended for benchmarks/tests only.

**Default Value**: 

.. code-block:: yaml

  "false"
  

level (int)
------------------------------------------------------------------------------------------------------------------------

Sets the minimum logging level.

**Default Value**: 

.. code-block:: yaml

  "3"
  

formatter (`logger.FormatterConfig`_)
------------------------------------------------------------------------------------------------------------------------

Sets logging format.

**Default Value**: 

.. code-block:: yaml

  type: json
  

logger.FormatterConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

type (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Sets logging format type.

**Default Value**: 

.. code-block:: yaml

  json
  

Section: otel
========================================================================================================================

type (string)
------------------------------------------------------------------------------------------------------------------------

Sets the type of exporter to configure [noop/file/jaeger].

**Default Value**: 

.. code-block:: yaml

  noop
  

file (`otelutils.FileConfig`_)
------------------------------------------------------------------------------------------------------------------------

Configuration for exporting telemetry traces to a file

**Default Value**: 

.. code-block:: yaml

  filename: /tmp/trace.txt
  

jaeger (`otelutils.JaegerConfig`_)
------------------------------------------------------------------------------------------------------------------------

Configuration for exporting telemetry traces to a jaeger

**Default Value**: 

.. code-block:: yaml

  endpoint: http://localhost:14268/api/traces
  

otelutils.FileConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

filename (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Filename to store exported telemetry traces

**Default Value**: 

.. code-block:: yaml

  /tmp/trace.txt
  

otelutils.JaegerConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

endpoint (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Endpoint for the jaeger telemtry trace ingestor

**Default Value**: 

.. code-block:: yaml

  http://localhost:14268/api/traces
  

Section: plugins
========================================================================================================================

agent-service (`agent.Config`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  agentForTaskTypes: null
  agents: null
  defaultAgent:
    defaultServiceConfig: ""
    defaultTimeout: 10s
    endpoint: ""
    insecure: true
    timeouts: null
  resourceConstraints:
    NamespaceScopeResourceConstraint:
      Value: 50
    ProjectScopeResourceConstraint:
      Value: 100
  supportedTaskTypes:
  - task_type_1
  - task_type_2
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
  

athena (`athena.Config`_)
------------------------------------------------------------------------------------------------------------------------

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
  

aws (`aws.Config`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  accountId: ""
  logLevel: 0
  region: us-east-2
  retries: 3
  

bigquery (`bigquery.Config`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  googleTokenSource:
    gke-task-workload-identity:
      remoteClusterConfig:
        auth:
          caCertPath: ""
          tokenPath: ""
        enabled: false
        endpoint: ""
        name: ""
    type: default
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
  

catalogcache (`catalog.Config`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  reader:
    maxItems: 10000
    maxRetries: 3
    workers: 10
  writer:
    maxItems: 10000
    maxRetries: 3
    workers: 10
  

databricks (`databricks.Config`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  databricksInstance: ""
  databricksTokenKey: FLYTE_DATABRICKS_API_TOKEN
  defaultWarehouse: COMPUTE_CLUSTER
  entrypointFile: ""
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
  

echo (`testing.Config`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  sleep-duration: 0s
  

k8s (`config.K8sPluginConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  co-pilot:
    cpu: 500m
    default-input-path: /var/flyte/inputs
    default-output-path: /var/flyte/outputs
    image: cr.flyte.org/flyteorg/flytecopilot:v0.0.15
    input-vol-name: flyte-inputs
    memory: 128Mi
    name: flyte-copilot-
    output-vol-name: flyte-outputs
    start-timeout: 1m40s
    storage: ""
  create-container-config-error-grace-period: 0s
  create-container-error-grace-period: 3m0s
  default-annotations:
    cluster-autoscaler.kubernetes.io/safe-to-evict: "false"
  default-cpus: "1"
  default-env-vars: null
  default-env-vars-from-env: null
  default-labels: null
  default-memory: 1Gi
  default-node-selector: null
  default-pod-dns-config: null
  default-pod-security-context: null
  default-pod-template-name: ""
  default-pod-template-resync: 30s
  default-security-context: null
  default-tolerations: null
  delete-resource-on-finalize: false
  enable-host-networking-pod: null
  gpu-device-node-label: k8s.amazonaws.com/accelerator
  gpu-partition-size-node-label: k8s.amazonaws.com/gpu-partition-size
  gpu-resource-name: nvidia.com/gpu
  gpu-unpartitioned-node-selector-requirement: null
  gpu-unpartitioned-toleration: null
  image-pull-backoff-grace-period: 3m0s
  inject-finalizer: false
  interruptible-node-selector: null
  interruptible-node-selector-requirement: null
  interruptible-tolerations: null
  non-interruptible-node-selector-requirement: null
  pod-pending-timeout: 0s
  resource-tolerations: null
  scheduler-name: ""
  send-object-events: false
  

k8s-array (`k8s.Config`_)
------------------------------------------------------------------------------------------------------------------------

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
      dynamic-log-links: null
      gcp-project: ""
      kubernetes-enabled: true
      kubernetes-template-uri: http://localhost:30082/#!/log/{{ .namespace }}/{{ .podName
        }}/pod?namespace={{ .namespace }}
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
  

kf-operator (`common.Config`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  timeout: 1m0s
  

logs (`logs.LogConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  cloudwatch-enabled: false
  cloudwatch-log-group: ""
  cloudwatch-region: ""
  cloudwatch-template-uri: ""
  dynamic-log-links: null
  gcp-project: ""
  kubernetes-enabled: true
  kubernetes-template-uri: http://localhost:30082/#!/log/{{ .namespace }}/{{ .podName
    }}/pod?namespace={{ .namespace }}
  kubernetes-url: ""
  stackdriver-enabled: false
  stackdriver-logresourcename: ""
  stackdriver-template-uri: ""
  templates: null
  

qubole (`config.Config`_)
------------------------------------------------------------------------------------------------------------------------

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
  

ray (`ray.Config`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  dashboardHost: 0.0.0.0
  dashboardURLTemplate: null
  defaults:
    headNode:
      ipAddress: $MY_POD_IP
      startParameters:
        disable-usage-stats: "true"
    workerNode:
      ipAddress: $MY_POD_IP
      startParameters:
        disable-usage-stats: "true"
  enableUsageStats: false
  includeDashboard: true
  logs:
    cloudwatch-enabled: false
    cloudwatch-log-group: ""
    cloudwatch-region: ""
    cloudwatch-template-uri: ""
    dynamic-log-links: null
    gcp-project: ""
    kubernetes-enabled: false
    kubernetes-template-uri: ""
    kubernetes-url: ""
    stackdriver-enabled: false
    stackdriver-logresourcename: ""
    stackdriver-template-uri: ""
    templates: null
  logsSidecar: null
  remoteClusterConfig:
    auth:
      caCertPath: ""
      tokenPath: ""
    enabled: false
    endpoint: ""
    name: ""
  serviceType: NodePort
  shutdownAfterJobFinishes: true
  ttlSecondsAfterFinished: 3600
  

snowflake (`snowflake.Config`_)
------------------------------------------------------------------------------------------------------------------------

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
  

spark (`spark.Config`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  features: null
  logs:
    all-user:
      cloudwatch-enabled: false
      cloudwatch-log-group: ""
      cloudwatch-region: ""
      cloudwatch-template-uri: ""
      dynamic-log-links: null
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
      dynamic-log-links: null
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
      dynamic-log-links: null
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
      dynamic-log-links: null
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
  

agent.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

webApi (`webapi.PluginConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines config for the base WebAPI plugin.

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
  

resourceConstraints (`core.ResourceConstraintsSpec`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  NamespaceScopeResourceConstraint:
    Value: 50
  ProjectScopeResourceConstraint:
    Value: 100
  

defaultAgent (`agent.Agent`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

The default agent.

**Default Value**: 

.. code-block:: yaml

  defaultServiceConfig: ""
  defaultTimeout: 10s
  endpoint: ""
  insecure: true
  timeouts: null
  

agents (map[string]*agent.Agent)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

The agents.

**Default Value**: 

.. code-block:: yaml

  null
  

agentForTaskTypes (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

supportedTaskTypes ([]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  - task_type_1
  - task_type_2
  

agent.Agent
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

endpoint (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

insecure (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "true"
  

defaultServiceConfig (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

timeouts (map[string]config.Duration)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

defaultTimeout (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  10s
  

core.ResourceConstraintsSpec
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

ProjectScopeResourceConstraint (`core.ResourceConstraint`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  Value: 100
  

NamespaceScopeResourceConstraint (`core.ResourceConstraint`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  Value: 50
  

core.ResourceConstraint
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Value (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "100"
  

webapi.PluginConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

resourceQuotas (webapi.ResourceQuotas)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  default: 1000
  

readRateLimiter (`webapi.RateLimiterConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines rate limiter properties for read actions (e.g. retrieve status).

**Default Value**: 

.. code-block:: yaml

  burst: 100
  qps: 10
  

writeRateLimiter (`webapi.RateLimiterConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines rate limiter properties for write actions.

**Default Value**: 

.. code-block:: yaml

  burst: 100
  qps: 10
  

caching (`webapi.CachingConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines caching characteristics.

**Default Value**: 

.. code-block:: yaml

  maxSystemFailures: 5
  resyncInterval: 30s
  size: 500000
  workers: 10
  

resourceMeta (interface)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  <nil>
  

webapi.CachingConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

size (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the maximum number of items to cache.

**Default Value**: 

.. code-block:: yaml

  "500000"
  

resyncInterval (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the sync interval.

**Default Value**: 

.. code-block:: yaml

  30s
  

workers (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the number of workers to start up to process items.

**Default Value**: 

.. code-block:: yaml

  "10"
  

maxSystemFailures (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the number of failures to fetch a task before failing the task.

**Default Value**: 

.. code-block:: yaml

  "5"
  

webapi.RateLimiterConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

qps (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the max rate of calls per second.

**Default Value**: 

.. code-block:: yaml

  "10"
  

burst (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the maximum burst size.

**Default Value**: 

.. code-block:: yaml

  "100"
  

athena.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

webApi (`webapi.PluginConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines config for the base WebAPI plugin.

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
  

resourceConstraints (`core.ResourceConstraintsSpec`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  NamespaceScopeResourceConstraint:
    Value: 50
  ProjectScopeResourceConstraint:
    Value: 100
  

defaultWorkGroup (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the default workgroup to use when running on Athena unless overwritten by the task.

**Default Value**: 

.. code-block:: yaml

  primary
  

defaultCatalog (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the default catalog to use when running on Athena unless overwritten by the task.

**Default Value**: 

.. code-block:: yaml

  AwsDataCatalog
  

aws.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

region (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

AWS Region to connect to.

**Default Value**: 

.. code-block:: yaml

  us-east-2
  

accountId (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

AWS Account Identifier.

**Default Value**: 

.. code-block:: yaml

  ""
  

retries (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Number of retries.

**Default Value**: 

.. code-block:: yaml

  "3"
  

logLevel (uint64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

bigquery.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

webApi (`webapi.PluginConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines config for the base WebAPI plugin.

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
  

resourceConstraints (`core.ResourceConstraintsSpec`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  NamespaceScopeResourceConstraint:
    Value: 50
  ProjectScopeResourceConstraint:
    Value: 100
  

googleTokenSource (`google.TokenSourceFactoryConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines Google token source

**Default Value**: 

.. code-block:: yaml

  gke-task-workload-identity:
    remoteClusterConfig:
      auth:
        caCertPath: ""
        tokenPath: ""
      enabled: false
      endpoint: ""
      name: ""
  type: default
  

bigQueryEndpoint (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

google.TokenSourceFactoryConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

type (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines type of TokenSourceFactory, possible values are 'default' and 'gke-task-workload-identity'

**Default Value**: 

.. code-block:: yaml

  default
  

gke-task-workload-identity (`google.GkeTaskWorkloadIdentityTokenSourceFactoryConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Extra configuration for GKE task workload identity token source factory

**Default Value**: 

.. code-block:: yaml

  remoteClusterConfig:
    auth:
      caCertPath: ""
      tokenPath: ""
    enabled: false
    endpoint: ""
    name: ""
  

google.GkeTaskWorkloadIdentityTokenSourceFactoryConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

remoteClusterConfig (`k8s.ClusterConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Configuration of remote GKE cluster

**Default Value**: 

.. code-block:: yaml

  auth:
    caCertPath: ""
    tokenPath: ""
  enabled: false
  endpoint: ""
  name: ""
  

k8s.ClusterConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

name (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Friendly name of the remote cluster

**Default Value**: 

.. code-block:: yaml

  ""
  

endpoint (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Remote K8s cluster endpoint

**Default Value**: 

.. code-block:: yaml

  ""
  

auth (`k8s.Auth`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  caCertPath: ""
  tokenPath: ""
  

enabled (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Boolean flag to enable or disable

**Default Value**: 

.. code-block:: yaml

  "false"
  

k8s.Auth
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

tokenPath (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Token path

**Default Value**: 

.. code-block:: yaml

  ""
  

caCertPath (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Certificate path

**Default Value**: 

.. code-block:: yaml

  ""
  

catalog.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

reader (`workqueue.Config`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Catalog reader workqueue config. Make sure the index cache must be big enough to accommodate the biggest array task allowed to run on the system.

**Default Value**: 

.. code-block:: yaml

  maxItems: 10000
  maxRetries: 3
  workers: 10
  

writer (`workqueue.Config`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Catalog writer workqueue config. Make sure the index cache must be big enough to accommodate the biggest array task allowed to run on the system.

**Default Value**: 

.. code-block:: yaml

  maxItems: 10000
  maxRetries: 3
  workers: 10
  

workqueue.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

workers (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Number of concurrent workers to start processing the queue.

**Default Value**: 

.. code-block:: yaml

  "10"
  

maxRetries (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum number of retries per item.

**Default Value**: 

.. code-block:: yaml

  "3"
  

maxItems (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum number of entries to keep in the index.

**Default Value**: 

.. code-block:: yaml

  "10000"
  

common.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

timeout (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  1m0s
  

config.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

endpoint (`config.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Endpoint for qubole to use

**Default Value**: 

.. code-block:: yaml

  https://wellness.qubole.com
  

commandApiPath (`config.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

API Path where commands can be launched on Qubole. Should be a valid url.

**Default Value**: 

.. code-block:: yaml

  /api/v1.2/commands/
  

analyzeLinkPath (`config.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

URL path where queries can be visualized on qubole website. Should be a valid url.

**Default Value**: 

.. code-block:: yaml

  /v2/analyze
  

quboleTokenKey (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Name of the key where to find Qubole token in the secret manager.

**Default Value**: 

.. code-block:: yaml

  FLYTE_QUBOLE_CLIENT_TOKEN
  

lruCacheSize (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Size of the AutoRefreshCache

**Default Value**: 

.. code-block:: yaml

  "2000"
  

workers (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Number of parallel workers to refresh the cache

**Default Value**: 

.. code-block:: yaml

  "15"
  

defaultClusterLabel (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

The default cluster label. This will be used if label is not specified on the hive job.

**Default Value**: 

.. code-block:: yaml

  default
  

clusterConfigs ([]config.ClusterConfig)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  - labels:
    - default
    limit: 100
    namespaceScopeQuotaProportionCap: 0.7
    primaryLabel: default
    projectScopeQuotaProportionCap: 0.7
  

destinationClusterConfigs ([]config.DestinationClusterConfig)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  []
  

config.K8sPluginConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

inject-finalizer (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Instructs the plugin to inject a finalizer on startTask and remove it on task termination.

**Default Value**: 

.. code-block:: yaml

  "false"
  

default-annotations (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  cluster-autoscaler.kubernetes.io/safe-to-evict: "false"
  

default-labels (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-env-vars (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-env-vars-from-env (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-cpus (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines a default value for cpu for containers if not specified.

**Default Value**: 

.. code-block:: yaml

  "1"
  

default-memory (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines a default value for memory for containers if not specified.

**Default Value**: 

.. code-block:: yaml

  1Gi
  

default-tolerations ([]v1.Toleration)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-node-selector (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-affinity (v1.Affinity)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

scheduler-name (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines scheduler name.

**Default Value**: 

.. code-block:: yaml

  ""
  

interruptible-tolerations ([]v1.Toleration)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

interruptible-node-selector (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

interruptible-node-selector-requirement (v1.NodeSelectorRequirement)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

non-interruptible-node-selector-requirement (v1.NodeSelectorRequirement)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

resource-tolerations (map[v1.ResourceName][]v1.Toleration)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

co-pilot (`config.FlyteCoPilotConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Co-Pilot Configuration

**Default Value**: 

.. code-block:: yaml

  cpu: 500m
  default-input-path: /var/flyte/inputs
  default-output-path: /var/flyte/outputs
  image: cr.flyte.org/flyteorg/flytecopilot:v0.0.15
  input-vol-name: flyte-inputs
  memory: 128Mi
  name: flyte-copilot-
  output-vol-name: flyte-outputs
  start-timeout: 1m40s
  storage: ""
  

delete-resource-on-finalize (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Instructs the system to delete the resource upon successful execution of a k8s pod rather than have the k8s garbage collector clean it up. This ensures that no resources are kept around (potentially consuming cluster resources). This, however, will cause k8s log links to expire as soon as the resource is finalized.

**Default Value**: 

.. code-block:: yaml

  "false"
  

create-container-error-grace-period (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  3m0s
  

create-container-config-error-grace-period (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  0s
  

image-pull-backoff-grace-period (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  3m0s
  

pod-pending-timeout (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  0s
  

gpu-device-node-label (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  k8s.amazonaws.com/accelerator
  

gpu-partition-size-node-label (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  k8s.amazonaws.com/gpu-partition-size
  

gpu-unpartitioned-node-selector-requirement (v1.NodeSelectorRequirement)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

gpu-unpartitioned-toleration (v1.Toleration)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

gpu-resource-name (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  nvidia.com/gpu
  

default-pod-security-context (v1.PodSecurityContext)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-security-context (v1.SecurityContext)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

enable-host-networking-pod (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  <invalid reflect.Value>
  

default-pod-dns-config (v1.PodDNSConfig)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-pod-template-name (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Name of the PodTemplate to use as the base for all k8s pods created by FlytePropeller.

**Default Value**: 

.. code-block:: yaml

  ""
  

default-pod-template-resync (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Frequency of resyncing default pod templates

**Default Value**: 

.. code-block:: yaml

  30s
  

send-object-events (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

If true, will send k8s object events in TaskExecutionEvent updates.

**Default Value**: 

.. code-block:: yaml

  "false"
  

config.FlyteCoPilotConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

name (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Flyte co-pilot sidecar container name prefix. (additional bits will be added after this)

**Default Value**: 

.. code-block:: yaml

  flyte-copilot-
  

image (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Flyte co-pilot Docker Image FQN

**Default Value**: 

.. code-block:: yaml

  cr.flyte.org/flyteorg/flytecopilot:v0.0.15
  

default-input-path (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Default path where the volume should be mounted

**Default Value**: 

.. code-block:: yaml

  /var/flyte/inputs
  

default-output-path (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Default path where the volume should be mounted

**Default Value**: 

.. code-block:: yaml

  /var/flyte/outputs
  

input-vol-name (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Name of the data volume that is created for storing inputs

**Default Value**: 

.. code-block:: yaml

  flyte-inputs
  

output-vol-name (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Name of the data volume that is created for storing outputs

**Default Value**: 

.. code-block:: yaml

  flyte-outputs
  

start-timeout (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  1m40s
  

cpu (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Used to set cpu for co-pilot containers

**Default Value**: 

.. code-block:: yaml

  500m
  

memory (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Used to set memory for co-pilot containers

**Default Value**: 

.. code-block:: yaml

  128Mi
  

storage (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Default storage limit for individual inputs / outputs

**Default Value**: 

.. code-block:: yaml

  ""
  

resource.Quantity
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

i (`resource.int64Amount`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  {}
  

d (`resource.infDecAmount`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  <nil>
  

s (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "1"
  

Format (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  DecimalSI
  

resource.infDecAmount
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Dec (inf.Dec)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

resource.int64Amount
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

value (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "1"
  

scale (int32)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

databricks.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

webApi (`webapi.PluginConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines config for the base WebAPI plugin.

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
  

resourceConstraints (`core.ResourceConstraintsSpec`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  NamespaceScopeResourceConstraint:
    Value: 50
  ProjectScopeResourceConstraint:
    Value: 100
  

defaultWarehouse (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the default warehouse to use when running on Databricks unless overwritten by the task.

**Default Value**: 

.. code-block:: yaml

  COMPUTE_CLUSTER
  

databricksTokenKey (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Name of the key where to find Databricks token in the secret manager.

**Default Value**: 

.. code-block:: yaml

  FLYTE_DATABRICKS_API_TOKEN
  

databricksInstance (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Databricks workspace instance name.

**Default Value**: 

.. code-block:: yaml

  ""
  

entrypointFile (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

A URL of the entrypoint file. DBFS and cloud storage (s3://, gcs://, adls://, etc) locations are supported.

**Default Value**: 

.. code-block:: yaml

  ""
  

databricksEndpoint (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

k8s.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

scheduler (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Decides the scheduler to use when launching array-pods.

**Default Value**: 

.. code-block:: yaml

  ""
  

maxErrorLength (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Determines the maximum length of the error string returned for the array.

**Default Value**: 

.. code-block:: yaml

  "1000"
  

maxArrayJobSize (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum size of array job.

**Default Value**: 

.. code-block:: yaml

  "5000"
  

resourceConfig (`k8s.ResourceConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  limit: 0
  primaryLabel: ""
  

remoteClusterConfig (`k8s.ClusterConfig (remoteClusterConfig)`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  auth:
    certPath: ""
    tokenPath: ""
    type: ""
  enabled: false
  endpoint: ""
  name: ""
  

node-selector (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

tolerations ([]v1.Toleration)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

namespaceTemplate (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

OutputAssembler (`workqueue.Config`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  maxItems: 100000
  maxRetries: 5
  workers: 10
  

ErrorAssembler (`workqueue.Config`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  maxItems: 100000
  maxRetries: 5
  workers: 10
  

logs (`k8s.LogConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Config for log links for k8s array jobs.

**Default Value**: 

.. code-block:: yaml

  config:
    cloudwatch-enabled: false
    cloudwatch-log-group: ""
    cloudwatch-region: ""
    cloudwatch-template-uri: ""
    dynamic-log-links: null
    gcp-project: ""
    kubernetes-enabled: true
    kubernetes-template-uri: http://localhost:30082/#!/log/{{ .namespace }}/{{ .podName
      }}/pod?namespace={{ .namespace }}
    kubernetes-url: ""
    stackdriver-enabled: false
    stackdriver-logresourcename: ""
    stackdriver-template-uri: ""
    templates: null
  

k8s.ClusterConfig (remoteClusterConfig)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

name (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Friendly name of the remote cluster

**Default Value**: 

.. code-block:: yaml

  ""
  

endpoint (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Remote K8s cluster endpoint

**Default Value**: 

.. code-block:: yaml

  ""
  

auth (`k8s.Auth (auth)`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  certPath: ""
  tokenPath: ""
  type: ""
  

enabled (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Boolean flag to enable or disable

**Default Value**: 

.. code-block:: yaml

  "false"
  

k8s.Auth (auth)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

type (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Authentication type

**Default Value**: 

.. code-block:: yaml

  ""
  

tokenPath (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Token path

**Default Value**: 

.. code-block:: yaml

  ""
  

certPath (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Certificate path

**Default Value**: 

.. code-block:: yaml

  ""
  

k8s.LogConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

config (`logs.LogConfig (config)`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the log config for k8s logs.

**Default Value**: 

.. code-block:: yaml

  cloudwatch-enabled: false
  cloudwatch-log-group: ""
  cloudwatch-region: ""
  cloudwatch-template-uri: ""
  dynamic-log-links: null
  gcp-project: ""
  kubernetes-enabled: true
  kubernetes-template-uri: http://localhost:30082/#!/log/{{ .namespace }}/{{ .podName
    }}/pod?namespace={{ .namespace }}
  kubernetes-url: ""
  stackdriver-enabled: false
  stackdriver-logresourcename: ""
  stackdriver-template-uri: ""
  templates: null
  

logs.LogConfig (config)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

cloudwatch-enabled (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Enable Cloudwatch Logging

**Default Value**: 

.. code-block:: yaml

  "false"
  

cloudwatch-region (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

AWS region in which Cloudwatch logs are stored.

**Default Value**: 

.. code-block:: yaml

  ""
  

cloudwatch-log-group (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Log group to which streams are associated.

**Default Value**: 

.. code-block:: yaml

  ""
  

cloudwatch-template-uri (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Template Uri to use when building cloudwatch log links

**Default Value**: 

.. code-block:: yaml

  ""
  

kubernetes-enabled (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Enable Kubernetes Logging

**Default Value**: 

.. code-block:: yaml

  "true"
  

kubernetes-url (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Console URL for Kubernetes logs

**Default Value**: 

.. code-block:: yaml

  ""
  

kubernetes-template-uri (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Template Uri to use when building kubernetes log links

**Default Value**: 

.. code-block:: yaml

  http://localhost:30082/#!/log/{{ .namespace }}/{{ .podName }}/pod?namespace={{ .namespace
    }}
  

stackdriver-enabled (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Enable Log-links to stackdriver

**Default Value**: 

.. code-block:: yaml

  "false"
  

gcp-project (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Name of the project in GCP

**Default Value**: 

.. code-block:: yaml

  ""
  

stackdriver-logresourcename (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Name of the logresource in stackdriver

**Default Value**: 

.. code-block:: yaml

  ""
  

stackdriver-template-uri (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Template Uri to use when building stackdriver log links

**Default Value**: 

.. code-block:: yaml

  ""
  

dynamic-log-links (map[string]tasklog.TemplateLogPlugin)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

templates ([]tasklog.TemplateLogPlugin)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

k8s.ResourceConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

primaryLabel (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

PrimaryLabel of a given service cluster

**Default Value**: 

.. code-block:: yaml

  ""
  

limit (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Resource quota (in the number of outstanding requests) for the cluster

**Default Value**: 

.. code-block:: yaml

  "0"
  

logs.LogConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

cloudwatch-enabled (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Enable Cloudwatch Logging

**Default Value**: 

.. code-block:: yaml

  "false"
  

cloudwatch-region (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

AWS region in which Cloudwatch logs are stored.

**Default Value**: 

.. code-block:: yaml

  ""
  

cloudwatch-log-group (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Log group to which streams are associated.

**Default Value**: 

.. code-block:: yaml

  ""
  

cloudwatch-template-uri (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Template Uri to use when building cloudwatch log links

**Default Value**: 

.. code-block:: yaml

  ""
  

kubernetes-enabled (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Enable Kubernetes Logging

**Default Value**: 

.. code-block:: yaml

  "true"
  

kubernetes-url (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Console URL for Kubernetes logs

**Default Value**: 

.. code-block:: yaml

  ""
  

kubernetes-template-uri (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Template Uri to use when building kubernetes log links

**Default Value**: 

.. code-block:: yaml

  http://localhost:30082/#!/log/{{ .namespace }}/{{ .podName }}/pod?namespace={{ .namespace
    }}
  

stackdriver-enabled (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Enable Log-links to stackdriver

**Default Value**: 

.. code-block:: yaml

  "false"
  

gcp-project (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Name of the project in GCP

**Default Value**: 

.. code-block:: yaml

  ""
  

stackdriver-logresourcename (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Name of the logresource in stackdriver

**Default Value**: 

.. code-block:: yaml

  ""
  

stackdriver-template-uri (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Template Uri to use when building stackdriver log links

**Default Value**: 

.. code-block:: yaml

  ""
  

dynamic-log-links (map[string]tasklog.TemplateLogPlugin)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

templates ([]tasklog.TemplateLogPlugin)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

ray.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

shutdownAfterJobFinishes (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "true"
  

ttlSecondsAfterFinished (int32)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "3600"
  

serviceType (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  NodePort
  

includeDashboard (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "true"
  

dashboardHost (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  0.0.0.0
  

nodeIPAddress (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

remoteClusterConfig (`k8s.ClusterConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Configuration of remote K8s cluster for ray jobs

**Default Value**: 

.. code-block:: yaml

  auth:
    caCertPath: ""
    tokenPath: ""
  enabled: false
  endpoint: ""
  name: ""
  

logs (`logs.LogConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  cloudwatch-enabled: false
  cloudwatch-log-group: ""
  cloudwatch-region: ""
  cloudwatch-template-uri: ""
  dynamic-log-links: null
  gcp-project: ""
  kubernetes-enabled: false
  kubernetes-template-uri: ""
  kubernetes-url: ""
  stackdriver-enabled: false
  stackdriver-logresourcename: ""
  stackdriver-template-uri: ""
  templates: null
  

logsSidecar (v1.Container)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

dashboardURLTemplate (tasklog.TemplateLogPlugin)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

defaults (`ray.DefaultConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  headNode:
    ipAddress: $MY_POD_IP
    startParameters:
      disable-usage-stats: "true"
  workerNode:
    ipAddress: $MY_POD_IP
    startParameters:
      disable-usage-stats: "true"
  

enableUsageStats (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Enable usage stats for ray jobs. These stats are submitted to usage-stats.ray.io per https://docs.ray.io/en/latest/cluster/usage-stats.html

**Default Value**: 

.. code-block:: yaml

  "false"
  

ray.DefaultConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

headNode (`ray.NodeConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ipAddress: $MY_POD_IP
  startParameters:
    disable-usage-stats: "true"
  

workerNode (`ray.NodeConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ipAddress: $MY_POD_IP
  startParameters:
    disable-usage-stats: "true"
  

ray.NodeConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

startParameters (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  disable-usage-stats: "true"
  

ipAddress (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  $MY_POD_IP
  

snowflake.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

webApi (`webapi.PluginConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines config for the base WebAPI plugin.

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
  

resourceConstraints (`core.ResourceConstraintsSpec`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  NamespaceScopeResourceConstraint:
    Value: 50
  ProjectScopeResourceConstraint:
    Value: 100
  

defaultWarehouse (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the default warehouse to use when running on Snowflake unless overwritten by the task.

**Default Value**: 

.. code-block:: yaml

  COMPUTE_WH
  

snowflakeTokenKey (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Name of the key where to find Snowflake token in the secret manager.

**Default Value**: 

.. code-block:: yaml

  FLYTE_SNOWFLAKE_CLIENT_TOKEN
  

snowflakeEndpoint (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

spark.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

spark-config-default (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

spark-history-server-url (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

URL for SparkHistory Server that each job will publish the execution history to.

**Default Value**: 

.. code-block:: yaml

  ""
  

features ([]spark.Feature)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

logs (`spark.LogConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Config for log links for spark applications.

**Default Value**: 

.. code-block:: yaml

  all-user:
    cloudwatch-enabled: false
    cloudwatch-log-group: ""
    cloudwatch-region: ""
    cloudwatch-template-uri: ""
    dynamic-log-links: null
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
    dynamic-log-links: null
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
    dynamic-log-links: null
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
    dynamic-log-links: null
    gcp-project: ""
    kubernetes-enabled: false
    kubernetes-template-uri: ""
    kubernetes-url: ""
    stackdriver-enabled: false
    stackdriver-logresourcename: ""
    stackdriver-template-uri: ""
    templates: null
  

spark.LogConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

mixed (`logs.LogConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the log config that's not split into user/system.

**Default Value**: 

.. code-block:: yaml

  cloudwatch-enabled: false
  cloudwatch-log-group: ""
  cloudwatch-region: ""
  cloudwatch-template-uri: ""
  dynamic-log-links: null
  gcp-project: ""
  kubernetes-enabled: true
  kubernetes-template-uri: http://localhost:30082/#!/log/{{ .namespace }}/{{ .podName
    }}/pod?namespace={{ .namespace }}
  kubernetes-url: ""
  stackdriver-enabled: false
  stackdriver-logresourcename: ""
  stackdriver-template-uri: ""
  templates: null
  

user (`logs.LogConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the log config for user logs.

**Default Value**: 

.. code-block:: yaml

  cloudwatch-enabled: false
  cloudwatch-log-group: ""
  cloudwatch-region: ""
  cloudwatch-template-uri: ""
  dynamic-log-links: null
  gcp-project: ""
  kubernetes-enabled: false
  kubernetes-template-uri: ""
  kubernetes-url: ""
  stackdriver-enabled: false
  stackdriver-logresourcename: ""
  stackdriver-template-uri: ""
  templates: null
  

system (`logs.LogConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the log config for system logs.

**Default Value**: 

.. code-block:: yaml

  cloudwatch-enabled: false
  cloudwatch-log-group: ""
  cloudwatch-region: ""
  cloudwatch-template-uri: ""
  dynamic-log-links: null
  gcp-project: ""
  kubernetes-enabled: false
  kubernetes-template-uri: ""
  kubernetes-url: ""
  stackdriver-enabled: false
  stackdriver-logresourcename: ""
  stackdriver-template-uri: ""
  templates: null
  

all-user (`logs.LogConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

All user logs across driver and executors.

**Default Value**: 

.. code-block:: yaml

  cloudwatch-enabled: false
  cloudwatch-log-group: ""
  cloudwatch-region: ""
  cloudwatch-template-uri: ""
  dynamic-log-links: null
  gcp-project: ""
  kubernetes-enabled: false
  kubernetes-template-uri: ""
  kubernetes-url: ""
  stackdriver-enabled: false
  stackdriver-logresourcename: ""
  stackdriver-template-uri: ""
  templates: null
  

testing.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

sleep-duration (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Indicates the amount of time before transitioning to success

**Default Value**: 

.. code-block:: yaml

  0s
  

Section: propeller
========================================================================================================================

kube-config (string)
------------------------------------------------------------------------------------------------------------------------

Path to kubernetes client config file.

**Default Value**: 

.. code-block:: yaml

  ""
  

master (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

workers (int)
------------------------------------------------------------------------------------------------------------------------

Number of threads to process workflows

**Default Value**: 

.. code-block:: yaml

  "20"
  

workflow-reeval-duration (`config.Duration`_)
------------------------------------------------------------------------------------------------------------------------

Frequency of re-evaluating workflows

**Default Value**: 

.. code-block:: yaml

  10s
  

downstream-eval-duration (`config.Duration`_)
------------------------------------------------------------------------------------------------------------------------

Frequency of re-evaluating downstream tasks

**Default Value**: 

.. code-block:: yaml

  30s
  

limit-namespace (string)
------------------------------------------------------------------------------------------------------------------------

Namespaces to watch for this propeller

**Default Value**: 

.. code-block:: yaml

  all
  

prof-port (`config.Port`_)
------------------------------------------------------------------------------------------------------------------------

Profiler port

**Default Value**: 

.. code-block:: yaml

  10254
  

metadata-prefix (string)
------------------------------------------------------------------------------------------------------------------------

MetadataPrefix should be used if all the metadata for Flyte executions should be stored under a specific prefix in CloudStorage. If not specified, the data will be stored in the base container directly.

**Default Value**: 

.. code-block:: yaml

  metadata/propeller
  

rawoutput-prefix (string)
------------------------------------------------------------------------------------------------------------------------

a fully qualified storage path of the form s3://flyte/abc/..., where all data sandboxes should be stored.

**Default Value**: 

.. code-block:: yaml

  ""
  

queue (`config.CompositeQueueConfig`_)
------------------------------------------------------------------------------------------------------------------------

Workflow workqueue configuration, affects the way the work is consumed from the queue.

**Default Value**: 

.. code-block:: yaml

  batch-size: -1
  batching-interval: 1s
  queue:
    base-delay: 0s
    capacity: 10000
    max-delay: 1m0s
    rate: 1000
    type: maxof
  sub-queue:
    base-delay: 0s
    capacity: 10000
    max-delay: 0s
    rate: 1000
    type: bucket
  type: batch
  

metrics-prefix (string)
------------------------------------------------------------------------------------------------------------------------

An optional prefix for all published metrics.

**Default Value**: 

.. code-block:: yaml

  flyte
  

metrics-keys ([]string)
------------------------------------------------------------------------------------------------------------------------

Metrics labels applied to prometheus metrics emitted by the service.

**Default Value**: 

.. code-block:: yaml

  - project
  - domain
  - wf
  - task
  

enable-admin-launcher (bool)
------------------------------------------------------------------------------------------------------------------------

Enable remote Workflow launcher to Admin

**Default Value**: 

.. code-block:: yaml

  "true"
  

max-workflow-retries (int)
------------------------------------------------------------------------------------------------------------------------

Maximum number of retries per workflow

**Default Value**: 

.. code-block:: yaml

  "10"
  

max-ttl-hours (int)
------------------------------------------------------------------------------------------------------------------------

Maximum number of hours a completed workflow should be retained. Number between 1-23 hours

**Default Value**: 

.. code-block:: yaml

  "23"
  

gc-interval (`config.Duration`_)
------------------------------------------------------------------------------------------------------------------------

Run periodic GC every 30 minutes

**Default Value**: 

.. code-block:: yaml

  30m0s
  

leader-election (`config.LeaderElectionConfig`_)
------------------------------------------------------------------------------------------------------------------------

Config for leader election.

**Default Value**: 

.. code-block:: yaml

  enabled: false
  lease-duration: 15s
  lock-config-map:
    Name: ""
    Namespace: ""
  renew-deadline: 10s
  retry-period: 2s
  

publish-k8s-events (bool)
------------------------------------------------------------------------------------------------------------------------

Enable events publishing to K8s events API.

**Default Value**: 

.. code-block:: yaml

  "false"
  

max-output-size-bytes (int64)
------------------------------------------------------------------------------------------------------------------------

Maximum size of outputs per task

**Default Value**: 

.. code-block:: yaml

  "10485760"
  

enable-grpc-latency-metrics (bool)
------------------------------------------------------------------------------------------------------------------------

Enable grpc latency metrics. Note Histograms metrics can be expensive on Prometheus servers.

**Default Value**: 

.. code-block:: yaml

  "false"
  

kube-client-config (`config.KubeClientConfig`_)
------------------------------------------------------------------------------------------------------------------------

Configuration to control the Kubernetes client

**Default Value**: 

.. code-block:: yaml

  burst: 25
  qps: 100
  timeout: 30s
  

node-config (`config.NodeConfig`_)
------------------------------------------------------------------------------------------------------------------------

config for a workflow node

**Default Value**: 

.. code-block:: yaml

  default-deadlines:
    node-active-deadline: 0s
    node-execution-deadline: 0s
    workflow-active-deadline: 0s
  default-max-attempts: 1
  enable-cr-debug-metadata: false
  ignore-retry-cause: false
  interruptible-failure-threshold: -1
  max-node-retries-system-failures: 3
  

max-streak-length (int)
------------------------------------------------------------------------------------------------------------------------

Maximum number of consecutive rounds that one propeller worker can use for one workflow - >1 => turbo-mode is enabled.

**Default Value**: 

.. code-block:: yaml

  "8"
  

event-config (`config.EventConfig`_)
------------------------------------------------------------------------------------------------------------------------

Configures execution event behavior.

**Default Value**: 

.. code-block:: yaml

  fallback-to-output-reference: false
  raw-output-policy: reference
  

include-shard-key-label ([]string)
------------------------------------------------------------------------------------------------------------------------

Include the specified shard key label in the k8s FlyteWorkflow CRD label selector

**Default Value**: 

.. code-block:: yaml

  []
  

exclude-shard-key-label ([]string)
------------------------------------------------------------------------------------------------------------------------

Exclude the specified shard key label from the k8s FlyteWorkflow CRD label selector

**Default Value**: 

.. code-block:: yaml

  []
  

include-project-label ([]string)
------------------------------------------------------------------------------------------------------------------------

Include the specified project label in the k8s FlyteWorkflow CRD label selector

**Default Value**: 

.. code-block:: yaml

  []
  

exclude-project-label ([]string)
------------------------------------------------------------------------------------------------------------------------

Exclude the specified project label from the k8s FlyteWorkflow CRD label selector

**Default Value**: 

.. code-block:: yaml

  []
  

include-domain-label ([]string)
------------------------------------------------------------------------------------------------------------------------

Include the specified domain label in the k8s FlyteWorkflow CRD label selector

**Default Value**: 

.. code-block:: yaml

  []
  

exclude-domain-label ([]string)
------------------------------------------------------------------------------------------------------------------------

Exclude the specified domain label from the k8s FlyteWorkflow CRD label selector

**Default Value**: 

.. code-block:: yaml

  []
  

cluster-id (string)
------------------------------------------------------------------------------------------------------------------------

Unique cluster id running this flytepropeller instance with which to annotate execution events

**Default Value**: 

.. code-block:: yaml

  propeller
  

create-flyteworkflow-crd (bool)
------------------------------------------------------------------------------------------------------------------------

Enable creation of the FlyteWorkflow CRD on startup

**Default Value**: 

.. code-block:: yaml

  "false"
  

array-node-event-version (int)
------------------------------------------------------------------------------------------------------------------------

ArrayNode eventing version. 0 => legacy (drop-in replacement for maptask), 1 => new

**Default Value**: 

.. code-block:: yaml

  "0"
  

node-execution-worker-count (int)
------------------------------------------------------------------------------------------------------------------------

Number of workers to evaluate node executions, currently only used for array nodes

**Default Value**: 

.. code-block:: yaml

  "8"
  

admin-launcher (`launchplan.AdminConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  burst: 10
  cacheSize: 10000
  tps: 100
  workers: 10
  

resourcemanager (`config.Config (resourcemanager)`_)
------------------------------------------------------------------------------------------------------------------------

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
  

workflowstore (`workflowstore.Config`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  policy: ResourceVersionCache
  

config.CompositeQueueConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

type (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Type of composite queue to use for the WorkQueue

**Default Value**: 

.. code-block:: yaml

  batch
  

queue (`config.WorkqueueConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Workflow workqueue configuration, affects the way the work is consumed from the queue.

**Default Value**: 

.. code-block:: yaml

  base-delay: 0s
  capacity: 10000
  max-delay: 1m0s
  rate: 1000
  type: maxof
  

sub-queue (`config.WorkqueueConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

SubQueue configuration, affects the way the nodes cause the top-level Work to be re-evaluated.

**Default Value**: 

.. code-block:: yaml

  base-delay: 0s
  capacity: 10000
  max-delay: 0s
  rate: 1000
  type: bucket
  

batching-interval (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Duration for which downstream updates are buffered

**Default Value**: 

.. code-block:: yaml

  1s
  

batch-size (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "-1"
  

config.WorkqueueConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

type (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Type of RateLimiter to use for the WorkQueue

**Default Value**: 

.. code-block:: yaml

  maxof
  

base-delay (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

base backoff delay for failure

**Default Value**: 

.. code-block:: yaml

  0s
  

max-delay (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Max backoff delay for failure

**Default Value**: 

.. code-block:: yaml

  1m0s
  

rate (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Bucket Refill rate per second

**Default Value**: 

.. code-block:: yaml

  "1000"
  

capacity (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Bucket capacity as number of items

**Default Value**: 

.. code-block:: yaml

  "10000"
  

config.Config (resourcemanager)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

type (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Which resource manager to use

**Default Value**: 

.. code-block:: yaml

  noop
  

resourceMaxQuota (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Global limit for concurrent Qubole queries

**Default Value**: 

.. code-block:: yaml

  "1000"
  

redis (`config.RedisConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Config for Redis resourcemanager.

**Default Value**: 

.. code-block:: yaml

  hostKey: ""
  hostPath: ""
  hostPaths: []
  maxRetries: 0
  primaryName: ""
  

config.RedisConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

hostPaths ([]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Redis hosts locations.

**Default Value**: 

.. code-block:: yaml

  []
  

primaryName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Redis primary name, fill in only if you are connecting to a redis sentinel cluster.

**Default Value**: 

.. code-block:: yaml

  ""
  

hostPath (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Redis host location

**Default Value**: 

.. code-block:: yaml

  ""
  

hostKey (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Key for local Redis access

**Default Value**: 

.. code-block:: yaml

  ""
  

maxRetries (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

See Redis client options for more info

**Default Value**: 

.. code-block:: yaml

  "0"
  

config.EventConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

raw-output-policy (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

How output data should be passed along in execution events.

**Default Value**: 

.. code-block:: yaml

  reference
  

fallback-to-output-reference (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Whether output data should be sent by reference when it is too large to be sent inline in execution events.

**Default Value**: 

.. code-block:: yaml

  "false"
  

config.KubeClientConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

qps (float32)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "100"
  

burst (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Max burst rate for throttle. 0 defaults to 10

**Default Value**: 

.. code-block:: yaml

  "25"
  

timeout (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Max duration allowed for every request to KubeAPI before giving up. 0 implies no timeout.

**Default Value**: 

.. code-block:: yaml

  30s
  

config.LeaderElectionConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

enabled (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Enables/Disables leader election.

**Default Value**: 

.. code-block:: yaml

  "false"
  

lock-config-map (`types.NamespacedName`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

ConfigMap namespace/name to use for resource lock.

**Default Value**: 

.. code-block:: yaml

  Name: ""
  Namespace: ""
  

lease-duration (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Duration that non-leader candidates will wait to force acquire leadership. This is measured against time of last observed ack.

**Default Value**: 

.. code-block:: yaml

  15s
  

renew-deadline (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Duration that the acting master will retry refreshing leadership before giving up.

**Default Value**: 

.. code-block:: yaml

  10s
  

retry-period (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Duration the LeaderElector clients should wait between tries of actions.

**Default Value**: 

.. code-block:: yaml

  2s
  

types.NamespacedName
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Namespace (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Name (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

config.NodeConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

default-deadlines (`config.DefaultDeadlines`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Default value for timeouts

**Default Value**: 

.. code-block:: yaml

  node-active-deadline: 0s
  node-execution-deadline: 0s
  workflow-active-deadline: 0s
  

max-node-retries-system-failures (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum number of retries per node for node failure due to infra issues

**Default Value**: 

.. code-block:: yaml

  "3"
  

interruptible-failure-threshold (int32)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

number of failures for a node to be still considered interruptible. Negative numbers are treated as complementary (ex. -1 means last attempt is non-interruptible).'

**Default Value**: 

.. code-block:: yaml

  "-1"
  

default-max-attempts (int32)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Default maximum number of attempts for a node

**Default Value**: 

.. code-block:: yaml

  "1"
  

ignore-retry-cause (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Ignore retry cause and count all attempts toward a node's max attempts

**Default Value**: 

.. code-block:: yaml

  "false"
  

enable-cr-debug-metadata (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Collapse node on any terminal state, not just successful terminations. This is useful to reduce the size of workflow state in etcd.

**Default Value**: 

.. code-block:: yaml

  "false"
  

config.DefaultDeadlines
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

node-execution-deadline (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Default value of node execution timeout that includes the time spent to run the node/workflow

**Default Value**: 

.. code-block:: yaml

  0s
  

node-active-deadline (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Default value of node timeout that includes the time spent queued.

**Default Value**: 

.. code-block:: yaml

  0s
  

workflow-active-deadline (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Default value of workflow timeout that includes the time spent queued.

**Default Value**: 

.. code-block:: yaml

  0s
  

config.Port
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

port (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "10254"
  

launchplan.AdminConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

tps (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

The maximum number of transactions per second to flyte admin from this client.

**Default Value**: 

.. code-block:: yaml

  "100"
  

burst (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum burst for throttle

**Default Value**: 

.. code-block:: yaml

  "10"
  

cacheSize (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum cache in terms of number of items stored.

**Default Value**: 

.. code-block:: yaml

  "10000"
  

workers (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Number of parallel workers to work on the queue.

**Default Value**: 

.. code-block:: yaml

  "10"
  

workflowstore.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

policy (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Workflow Store Policy to initialize

**Default Value**: 

.. code-block:: yaml

  ResourceVersionCache
  

Section: secrets
========================================================================================================================

secrets-prefix (string)
------------------------------------------------------------------------------------------------------------------------

Prefix where to look for secrets file

**Default Value**: 

.. code-block:: yaml

  /etc/secrets
  

env-prefix (string)
------------------------------------------------------------------------------------------------------------------------

Prefix for environment variables

**Default Value**: 

.. code-block:: yaml

  FLYTE_SECRET_
  

Section: storage
========================================================================================================================

type (string)
------------------------------------------------------------------------------------------------------------------------

Sets the type of storage to configure [s3/minio/local/mem/stow].

**Default Value**: 

.. code-block:: yaml

  s3
  

connection (`storage.ConnectionConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  access-key: ""
  auth-type: iam
  disable-ssl: false
  endpoint: ""
  region: us-east-1
  secret-key: ""
  

stow (`storage.StowConfig`_)
------------------------------------------------------------------------------------------------------------------------

Storage config for stow backend.

**Default Value**: 

.. code-block:: yaml

  {}
  

container (string)
------------------------------------------------------------------------------------------------------------------------

Initial container (in s3 a bucket) to create -if it doesn't exist-.'

**Default Value**: 

.. code-block:: yaml

  ""
  

enable-multicontainer (bool)
------------------------------------------------------------------------------------------------------------------------

If this is true, then the container argument is overlooked and redundant. This config will automatically open new connections to new containers/buckets as they are encountered

**Default Value**: 

.. code-block:: yaml

  "false"
  

cache (`storage.CachingConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  max_size_mbs: 0
  target_gc_percent: 0
  

limits (`storage.LimitsConfig`_)
------------------------------------------------------------------------------------------------------------------------

Sets limits for stores.

**Default Value**: 

.. code-block:: yaml

  maxDownloadMBs: 2
  

defaultHttpClient (`storage.HTTPClientConfig`_)
------------------------------------------------------------------------------------------------------------------------

Sets the default http client config.

**Default Value**: 

.. code-block:: yaml

  headers: null
  timeout: 0s
  

signedUrl (`storage.SignedURLConfig`_)
------------------------------------------------------------------------------------------------------------------------

Sets config for SignedURL.

**Default Value**: 

.. code-block:: yaml

  {}
  

storage.CachingConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

max_size_mbs (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum size of the cache where the Blob store data is cached in-memory. If not specified or set to 0, cache is not used

**Default Value**: 

.. code-block:: yaml

  "0"
  

target_gc_percent (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Sets the garbage collection target percentage.

**Default Value**: 

.. code-block:: yaml

  "0"
  

storage.ConnectionConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

endpoint (`config.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

URL for storage client to connect to.

**Default Value**: 

.. code-block:: yaml

  ""
  

auth-type (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Auth Type to use [iam,accesskey].

**Default Value**: 

.. code-block:: yaml

  iam
  

access-key (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Access key to use. Only required when authtype is set to accesskey.

**Default Value**: 

.. code-block:: yaml

  ""
  

secret-key (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Secret to use when accesskey is set.

**Default Value**: 

.. code-block:: yaml

  ""
  

region (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Region to connect to.

**Default Value**: 

.. code-block:: yaml

  us-east-1
  

disable-ssl (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Disables SSL connection. Should only be used for development.

**Default Value**: 

.. code-block:: yaml

  "false"
  

storage.HTTPClientConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

headers (map[string][]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

timeout (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Sets time out on the http client.

**Default Value**: 

.. code-block:: yaml

  0s
  

storage.LimitsConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

maxDownloadMBs (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum allowed download size (in MBs) per call.

**Default Value**: 

.. code-block:: yaml

  "2"
  

storage.SignedURLConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

stowConfigOverride (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

storage.StowConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

kind (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Kind of Stow backend to use. Refer to github/flyteorg/stow

**Default Value**: 

.. code-block:: yaml

  ""
  

config (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Configuration for stow backend. Refer to github/flyteorg/stow

**Default Value**: 

.. code-block:: yaml

  {}
  

Section: tasks
========================================================================================================================

task-plugins (`config.TaskPluginConfig`_)
------------------------------------------------------------------------------------------------------------------------

Task plugin configuration

**Default Value**: 

.. code-block:: yaml

  default-for-task-types: {}
  enabled-plugins: []
  

max-plugin-phase-versions (int32)
------------------------------------------------------------------------------------------------------------------------

Maximum number of plugin phase versions allowed for one phase.

**Default Value**: 

.. code-block:: yaml

  "100000"
  

backoff (`config.BackOffConfig`_)
------------------------------------------------------------------------------------------------------------------------

Config for Exponential BackOff implementation

**Default Value**: 

.. code-block:: yaml

  base-second: 2
  max-duration: 20s
  

maxLogMessageLength (int)
------------------------------------------------------------------------------------------------------------------------

Deprecated!!! Max length of error message.

**Default Value**: 

.. code-block:: yaml

  "0"
  

config.BackOffConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

base-second (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

The number of seconds representing the base duration of the exponential backoff

**Default Value**: 

.. code-block:: yaml

  "2"
  

max-duration (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

The cap of the backoff duration

**Default Value**: 

.. code-block:: yaml

  20s
  

config.TaskPluginConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

enabled-plugins ([]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Plugins enabled currently

**Default Value**: 

.. code-block:: yaml

  []
  

default-for-task-types (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  {}
  

Section: webhook
========================================================================================================================

metrics-prefix (string)
------------------------------------------------------------------------------------------------------------------------

An optional prefix for all published metrics.

**Default Value**: 

.. code-block:: yaml

  'flyte:'
  

certDir (string)
------------------------------------------------------------------------------------------------------------------------

Certificate directory to use to write generated certs. Defaults to /etc/webhook/certs/

**Default Value**: 

.. code-block:: yaml

  /etc/webhook/certs
  

localCert (bool)
------------------------------------------------------------------------------------------------------------------------

write certs locally. Defaults to false

**Default Value**: 

.. code-block:: yaml

  "false"
  

listenPort (int)
------------------------------------------------------------------------------------------------------------------------

The port to use to listen to webhook calls. Defaults to 9443

**Default Value**: 

.. code-block:: yaml

  "9443"
  

serviceName (string)
------------------------------------------------------------------------------------------------------------------------

The name of the webhook service.

**Default Value**: 

.. code-block:: yaml

  flyte-pod-webhook
  

servicePort (int32)
------------------------------------------------------------------------------------------------------------------------

The port on the service that hosting webhook.

**Default Value**: 

.. code-block:: yaml

  "443"
  

secretName (string)
------------------------------------------------------------------------------------------------------------------------

Secret name to write generated certs to.

**Default Value**: 

.. code-block:: yaml

  flyte-pod-webhook
  

secretManagerType (int)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  K8s
  

awsSecretManager (`config.AWSSecretManagerConfig`_)
------------------------------------------------------------------------------------------------------------------------

AWS Secret Manager config.

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
  

gcpSecretManager (`config.GCPSecretManagerConfig`_)
------------------------------------------------------------------------------------------------------------------------

GCP Secret Manager config.

**Default Value**: 

.. code-block:: yaml

  resources:
    limits:
      cpu: 200m
      memory: 500Mi
    requests:
      cpu: 200m
      memory: 500Mi
  sidecarImage: gcr.io/google.com/cloudsdktool/cloud-sdk:alpine
  

vaultSecretManager (`config.VaultSecretManagerConfig`_)
------------------------------------------------------------------------------------------------------------------------

Vault Secret Manager config.

**Default Value**: 

.. code-block:: yaml

  annotations: null
  kvVersion: "2"
  role: flyte
  

config.AWSSecretManagerConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

sidecarImage (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Specifies the sidecar docker image to use

**Default Value**: 

.. code-block:: yaml

  docker.io/amazon/aws-secrets-manager-secret-sidecar:v0.1.4
  

resources (`v1.ResourceRequirements`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  limits:
    cpu: 200m
    memory: 500Mi
  requests:
    cpu: 200m
    memory: 500Mi
  

v1.ResourceRequirements
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

limits (v1.ResourceList)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  cpu: 200m
  memory: 500Mi
  

requests (v1.ResourceList)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  cpu: 200m
  memory: 500Mi
  

claims ([]v1.ResourceClaim)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

config.GCPSecretManagerConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

sidecarImage (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Specifies the sidecar docker image to use

**Default Value**: 

.. code-block:: yaml

  gcr.io/google.com/cloudsdktool/cloud-sdk:alpine
  

resources (`v1.ResourceRequirements`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  limits:
    cpu: 200m
    memory: 500Mi
  requests:
    cpu: 200m
    memory: 500Mi
  

config.VaultSecretManagerConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

role (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Specifies the vault role to use

**Default Value**: 

.. code-block:: yaml

  flyte
  

kvVersion (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "2"
  

annotations (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

