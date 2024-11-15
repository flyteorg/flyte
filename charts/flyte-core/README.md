# flyte-core

![Version: v0.1.10](https://img.shields.io/badge/Version-v0.1.10-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)

A Helm chart for Flyte core

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../flyteagent | flyteagent(flyteagent) | v0.1.10 |

### Flyte INSTALLATION:
- [Install helm 3](https://helm.sh/docs/intro/install/)
- Fetch chart dependencies ``
- Install Flyte:

```bash
helm repo add flyte https://flyteorg.github.io/flyte
helm install -n flyte -f values-eks.yaml --create-namespace flyte flyte/flyte-core
```

Customize your installation by changing settings in `values-eks.yaml`.
You can use the helm diff plugin to review any value changes you've made to your values:

```bash
helm plugin install https://github.com/databus23/helm-diff
helm diff upgrade -f values-eks.yaml flyte flyte/flyte-core
```

Then apply your changes:
```bash
helm upgrade -f values-eks.yaml flyte flyte/flyte-core
```

Install ingress controller (By default Flyte helm chart have contour ingress resource)
```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install gateway bitnami/contour -n flyte
```

#### Alternative: Generate raw kubernetes yaml with helm template
- `helm template --name-template=flyte-eks . -n flyte -f values-eks.yaml > flyte_generated_eks.yaml`
- Deploy the manifest `kubectl apply -f flyte_generated_eks.yaml`

- When all pods are running - run end2end tests: `kubectl apply -f ../end2end/tests/endtoend.yaml`
- Get flyte host `minikube service contour -n heptio-contour --url`. And then visit `http://<HOST>/console`

### CONFIGURATION NOTES:
- The docker images, their tags and other default parameters are configured in `values.yaml` file.
- Each Flyte installation type should have separate `values-*.yaml` file: for sandbox, EKS and etc. The configuration in `values.yaml` and the chosen config `values-*.yaml` are merged when generating the deployment manifest.
- The configuration in `values-sandbox.yaml` is ready for installation in minikube. But `values-eks.yaml` should be edited before installation: s3 bucket, RDS hosts, iam roles, secrets and etc need to be modified.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| cloud_events.aws | object | `{"region":"us-east-2"}` | Configuration for sending cloud events to AWS SNS |
| cloud_events.enable | bool | `false` |  |
| cloud_events.eventsPublisher.eventTypes[0] | string | `"all"` |  |
| cloud_events.eventsPublisher.topicName | string | `"arn:aws:sns:us-east-2:123456:123-my-topic"` |  |
| cloud_events.gcp | object | `{"region":"us-east1"}` | Configuration for sending cloud events to GCP Pub Sub |
| cloud_events.kafka | object | `{"brokers":["mybroker:443"],"saslConfig":{"enabled":false,"handshake":true,"mechanism":"PLAIN","password":"","passwordPath":"","user":"kafka"},"tlsConfig":{"certPath":"/etc/ssl/certs/kafka-client.crt","enabled":false,"keyPath":"/etc/ssl/certs/kafka-client.key"},"version":"3.7.0"}` | Configuration for sending cloud events to Kafka |
| cloud_events.kafka.brokers | list | `["mybroker:443"]` | The kafka brokers to talk to |
| cloud_events.kafka.saslConfig | object | `{"enabled":false,"handshake":true,"mechanism":"PLAIN","password":"","passwordPath":"","user":"kafka"}` | SASL based authentication |
| cloud_events.kafka.saslConfig.enabled | bool | `false` | Whether to use SASL authentication |
| cloud_events.kafka.saslConfig.handshake | bool | `true` | Whether the send the SASL handsahke first |
| cloud_events.kafka.saslConfig.mechanism | string | `"PLAIN"` | Which SASL mechanism to use. Defaults to PLAIN |
| cloud_events.kafka.saslConfig.password | string | `""` | The password for the kafka user |
| cloud_events.kafka.saslConfig.passwordPath | string | `""` | Optional mount path of file containing the kafka password. |
| cloud_events.kafka.saslConfig.user | string | `"kafka"` | The kafka user |
| cloud_events.kafka.tlsConfig | object | `{"certPath":"/etc/ssl/certs/kafka-client.crt","enabled":false,"keyPath":"/etc/ssl/certs/kafka-client.key"}` | Certificate based authentication |
| cloud_events.kafka.tlsConfig.certPath | string | `"/etc/ssl/certs/kafka-client.crt"` | Path to the client certificate |
| cloud_events.kafka.tlsConfig.enabled | bool | `false` | Whether to use certificate based authentication or TLS |
| cloud_events.kafka.tlsConfig.keyPath | string | `"/etc/ssl/certs/kafka-client.key"` | Path to the client private key |
| cloud_events.kafka.version | string | `"3.7.0"` | The version of Kafka |
| cloud_events.type | string | `"aws"` |  |
| cluster_resource_manager | object | `{"config":{"cluster_resources":{"customData":[{"production":[{"projectQuotaCpu":{"value":"5"}},{"projectQuotaMemory":{"value":"4000Mi"}}]},{"staging":[{"projectQuotaCpu":{"value":"2"}},{"projectQuotaMemory":{"value":"3000Mi"}}]},{"development":[{"projectQuotaCpu":{"value":"4"}},{"projectQuotaMemory":{"value":"3000Mi"}}]}],"refreshInterval":"5m","standaloneDeployment":false,"templatePath":"/etc/flyte/clusterresource/templates"}},"enabled":true,"nodeSelector":{},"podAnnotations":{},"podEnv":{},"podLabels":{},"prometheus":{"enabled":false,"path":"/metrics","port":10254},"resources":{},"service_account_name":"flyteadmin","standaloneDeployment":false,"templates":[{"key":"aa_namespace","value":"apiVersion: v1\nkind: Namespace\nmetadata:\n  name: {{ namespace }}\nspec:\n  finalizers:\n  - kubernetes\n"},{"key":"ab_project_resource_quota","value":"apiVersion: v1\nkind: ResourceQuota\nmetadata:\n  name: project-quota\n  namespace: {{ namespace }}\nspec:\n  hard:\n    limits.cpu: {{ projectQuotaCpu }}\n    limits.memory: {{ projectQuotaMemory }}\n"}]}` | Configuration for the Cluster resource manager component. This is an optional component, that enables automatic cluster configuration. This is useful to set default quotas, manage namespaces etc that map to a project/domain |
| cluster_resource_manager.config | object | `{"cluster_resources":{"customData":[{"production":[{"projectQuotaCpu":{"value":"5"}},{"projectQuotaMemory":{"value":"4000Mi"}}]},{"staging":[{"projectQuotaCpu":{"value":"2"}},{"projectQuotaMemory":{"value":"3000Mi"}}]},{"development":[{"projectQuotaCpu":{"value":"4"}},{"projectQuotaMemory":{"value":"3000Mi"}}]}],"refreshInterval":"5m","standaloneDeployment":false,"templatePath":"/etc/flyte/clusterresource/templates"}}` | Configmap for ClusterResource parameters |
| cluster_resource_manager.config.cluster_resources | object | `{"customData":[{"production":[{"projectQuotaCpu":{"value":"5"}},{"projectQuotaMemory":{"value":"4000Mi"}}]},{"staging":[{"projectQuotaCpu":{"value":"2"}},{"projectQuotaMemory":{"value":"3000Mi"}}]},{"development":[{"projectQuotaCpu":{"value":"4"}},{"projectQuotaMemory":{"value":"3000Mi"}}]}],"refreshInterval":"5m","standaloneDeployment":false,"templatePath":"/etc/flyte/clusterresource/templates"}` | ClusterResource parameters Refer to the [structure](https://pkg.go.dev/github.com/lyft/flyteadmin@v0.3.37/pkg/runtime/interfaces#ClusterResourceConfig) to customize. |
| cluster_resource_manager.config.cluster_resources.refreshInterval | string | `"5m"` | How frequently to run the sync process |
| cluster_resource_manager.config.cluster_resources.standaloneDeployment | bool | `false` | Starts the cluster resource manager in standalone mode with requisite auth credentials to call flyteadmin service endpoints |
| cluster_resource_manager.enabled | bool | `true` | Enables the Cluster resource manager component |
| cluster_resource_manager.nodeSelector | object | `{}` | nodeSelector for ClusterResource deployment |
| cluster_resource_manager.podAnnotations | object | `{}` | Annotations for ClusterResource pods |
| cluster_resource_manager.podEnv | object | `{}` | Additional ClusterResource container environment variables |
| cluster_resource_manager.podLabels | object | `{}` | Labels for ClusterResource pods |
| cluster_resource_manager.resources | object | `{}` | Resources for ClusterResource deployment |
| cluster_resource_manager.service_account_name | string | `"flyteadmin"` | Service account name to run with |
| cluster_resource_manager.templates | list | `[{"key":"aa_namespace","value":"apiVersion: v1\nkind: Namespace\nmetadata:\n  name: {{ namespace }}\nspec:\n  finalizers:\n  - kubernetes\n"},{"key":"ab_project_resource_quota","value":"apiVersion: v1\nkind: ResourceQuota\nmetadata:\n  name: project-quota\n  namespace: {{ namespace }}\nspec:\n  hard:\n    limits.cpu: {{ projectQuotaCpu }}\n    limits.memory: {{ projectQuotaMemory }}\n"}]` | Resource templates that should be applied |
| cluster_resource_manager.templates[0] | object | `{"key":"aa_namespace","value":"apiVersion: v1\nkind: Namespace\nmetadata:\n  name: {{ namespace }}\nspec:\n  finalizers:\n  - kubernetes\n"}` | Template for namespaces resources |
| common | object | `{"databaseSecret":{"name":"","secretManifest":{}},"flyteNamespaceTemplate":{"enabled":false},"ingress":{"albSSLRedirect":false,"annotations":{"nginx.ingress.kubernetes.io/app-root":"/console","nginx.ingress.kubernetes.io/service-upstream":"true"},"enabled":true,"ingressClassName":null,"separateGrpcIngress":false,"separateGrpcIngressAnnotations":{"nginx.ingress.kubernetes.io/backend-protocol":"GRPC"},"tls":{"enabled":false},"webpackHMR":false}}` | ----------------------------------------------  COMMON SETTINGS  |
| common.databaseSecret.name | string | `""` | Specify name of K8s Secret which contains Database password. Leave it empty if you don't need this Secret |
| common.databaseSecret.secretManifest | object | `{}` | Specify your Secret (with sensitive data) or pseudo-manifest (without sensitive data). See https://github.com/godaddy/kubernetes-external-secrets |
| common.flyteNamespaceTemplate.enabled | bool | `false` | - Enable or disable creating Flyte namespace in template. Enable when using helm as template-engine only. Disable when using `helm install ...`. |
| common.ingress.albSSLRedirect | bool | `false` | - albSSLRedirect adds a special route for ssl redirect. Only useful in combination with the AWS LoadBalancer Controller. |
| common.ingress.annotations | object | `{"nginx.ingress.kubernetes.io/app-root":"/console","nginx.ingress.kubernetes.io/service-upstream":"true"}` | - Ingress annotations applied to both HTTP and GRPC ingresses. |
| common.ingress.enabled | bool | `true` | - Enable or disable creating Ingress for Flyte. Relevant to disable when using e.g. Istio as ingress controller. |
| common.ingress.ingressClassName | string | `nil` | - Sets the ingressClassName |
| common.ingress.separateGrpcIngress | bool | `false` | - separateGrpcIngress puts GRPC routes into a separate ingress if true. Required for certain ingress controllers like nginx. |
| common.ingress.separateGrpcIngressAnnotations | object | `{"nginx.ingress.kubernetes.io/backend-protocol":"GRPC"}` | - Extra Ingress annotations applied only to the GRPC ingress. Only makes sense if `separateGrpcIngress` is enabled. |
| common.ingress.tls | object | `{"enabled":false}` | - Ingress hostname host: |
| common.ingress.webpackHMR | bool | `false` | - Enable or disable HMR route to flyteconsole. This is useful only for frontend development. |
| configmap.admin | object | `{"admin":{"clientId":"{{ .Values.secrets.adminOauthClientCredentials.clientId }}","clientSecretLocation":"/etc/secrets/client_secret","endpoint":"flyteadmin:81","insecure":true},"event":{"capacity":1000,"rate":500,"type":"admin"}}` | Admin Client configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/subworkflow/launchplan#AdminConfig) |
| configmap.adminServer | object | `{"auth":{"appAuth":{"thirdPartyConfig":{"flyteClient":{"clientId":"flytectl","redirectUri":"http://localhost:53593/callback","scopes":["offline","all"]}}},"authorizedUris":["https://localhost:30081","http://flyteadmin:80","http://flyteadmin.flyte.svc.cluster.local:80"],"userAuth":{"openId":{"baseUrl":"https://accounts.google.com","clientId":"657465813211-6eog7ek7li5k7i7fvgv2921075063hpe.apps.googleusercontent.com","scopes":["profile","openid"]}}},"flyteadmin":{"eventVersion":2,"metadataStoragePrefix":["metadata","admin"],"metricsScope":"flyte:","profilerPort":10254,"roleNameKey":"iam.amazonaws.com/role","testing":{"host":"http://flyteadmin"}},"server":{"grpc":{"port":8089},"httpPort":8088,"security":{"allowCors":true,"allowedHeaders":["Content-Type","flyte-authorization"],"allowedOrigins":["*"],"secure":false,"useAuth":false}}}` | FlyteAdmin server configuration |
| configmap.adminServer.auth | object | `{"appAuth":{"thirdPartyConfig":{"flyteClient":{"clientId":"flytectl","redirectUri":"http://localhost:53593/callback","scopes":["offline","all"]}}},"authorizedUris":["https://localhost:30081","http://flyteadmin:80","http://flyteadmin.flyte.svc.cluster.local:80"],"userAuth":{"openId":{"baseUrl":"https://accounts.google.com","clientId":"657465813211-6eog7ek7li5k7i7fvgv2921075063hpe.apps.googleusercontent.com","scopes":["profile","openid"]}}}` | Authentication configuration |
| configmap.adminServer.server.security.secure | bool | `false` | Controls whether to serve requests over SSL/TLS. |
| configmap.adminServer.server.security.useAuth | bool | `false` | Controls whether to enforce authentication. Follow the guide in https://docs.flyte.org/ on how to setup authentication. |
| configmap.catalog | object | `{"catalog-cache":{"endpoint":"datacatalog:89","insecure":true,"type":"datacatalog"}}` | Catalog Client configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/catalog#Config) Additional advanced Catalog configuration [here](https://pkg.go.dev/github.com/lyft/flyteplugins/go/tasks/pluginmachinery/catalog#Config) |
| configmap.clusters.clusterConfigs | list | `[]` |  |
| configmap.clusters.labelClusterMap | object | `{}` |  |
| configmap.console | object | `{"BASE_URL":"/console","CONFIG_DIR":"/etc/flyte/config"}` | Configuration for Flyte console UI |
| configmap.copilot | object | `{"plugins":{"k8s":{"co-pilot":{"image":"cr.flyte.org/flyteorg/flytecopilot:v1.13.2","name":"flyte-copilot-","start-timeout":"30s"}}}}` | Copilot configuration |
| configmap.copilot.plugins.k8s.co-pilot | object | `{"image":"cr.flyte.org/flyteorg/flytecopilot:v1.13.2","name":"flyte-copilot-","start-timeout":"30s"}` | Structure documented [here](https://pkg.go.dev/github.com/lyft/flyteplugins@v0.5.28/go/tasks/pluginmachinery/flytek8s/config#FlyteCoPilotConfig) |
| configmap.core | object | `{"manager":{"pod-application":"flytepropeller","pod-template-container-name":"flytepropeller","pod-template-name":"flytepropeller-template"},"propeller":{"downstream-eval-duration":"30s","enable-admin-launcher":true,"leader-election":{"enabled":true,"lease-duration":"15s","lock-config-map":{"name":"propeller-leader","namespace":"flyte"},"renew-deadline":"10s","retry-period":"2s"},"limit-namespace":"all","max-workflow-retries":30,"metadata-prefix":"metadata/propeller","metrics-prefix":"flyte","prof-port":10254,"queue":{"batch-size":-1,"batching-interval":"2s","queue":{"base-delay":"5s","capacity":1000,"max-delay":"120s","rate":100,"type":"maxof"},"sub-queue":{"capacity":100,"rate":10,"type":"bucket"},"type":"batch"},"rawoutput-prefix":"s3://my-s3-bucket/","workers":4,"workflow-reeval-duration":"30s"},"webhook":{"certDir":"/etc/webhook/certs","serviceName":"flyte-pod-webhook"}}` | Core propeller configuration |
| configmap.core.manager | object | `{"pod-application":"flytepropeller","pod-template-container-name":"flytepropeller","pod-template-name":"flytepropeller-template"}` | follows the structure specified [here](https://pkg.go.dev/github.com/flyteorg/flytepropeller/manager/config#Config). |
| configmap.core.propeller | object | `{"downstream-eval-duration":"30s","enable-admin-launcher":true,"leader-election":{"enabled":true,"lease-duration":"15s","lock-config-map":{"name":"propeller-leader","namespace":"flyte"},"renew-deadline":"10s","retry-period":"2s"},"limit-namespace":"all","max-workflow-retries":30,"metadata-prefix":"metadata/propeller","metrics-prefix":"flyte","prof-port":10254,"queue":{"batch-size":-1,"batching-interval":"2s","queue":{"base-delay":"5s","capacity":1000,"max-delay":"120s","rate":100,"type":"maxof"},"sub-queue":{"capacity":100,"rate":10,"type":"bucket"},"type":"batch"},"rawoutput-prefix":"s3://my-s3-bucket/","workers":4,"workflow-reeval-duration":"30s"}` | follows the structure specified [here](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/config). |
| configmap.datacatalogServer | object | `{"application":{"grpcPort":8089,"grpcServerReflection":true,"httpPort":8080},"datacatalog":{"heartbeat-grace-period-multiplier":3,"max-reservation-heartbeat":"30s","metrics-scope":"datacatalog","profiler-port":10254,"storage-prefix":"metadata/datacatalog"}}` | Datacatalog server config |
| configmap.domain | object | `{"domains":[{"id":"development","name":"development"},{"id":"staging","name":"staging"},{"id":"production","name":"production"}]}` | Domains configuration for Flyte projects. This enables the specified number of domains across all projects in Flyte. |
| configmap.enabled_plugins.tasks | object | `{"task-plugins":{"default-for-task-types":{"container":"container","container_array":"k8s-array","sidecar":"sidecar"},"enabled-plugins":["container","sidecar","k8s-array","agent-service","echo"]}}` | Tasks specific configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#GetConfig) |
| configmap.enabled_plugins.tasks.task-plugins | object | `{"default-for-task-types":{"container":"container","container_array":"k8s-array","sidecar":"sidecar"},"enabled-plugins":["container","sidecar","k8s-array","agent-service","echo"]}` | Plugins configuration, [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#TaskPluginConfig) |
| configmap.enabled_plugins.tasks.task-plugins.enabled-plugins | list | `["container","sidecar","k8s-array","agent-service","echo"]` | [Enabled Plugins](https://pkg.go.dev/github.com/lyft/flyteplugins/go/tasks/config#Config). Enable sagemaker*, athena if you install the backend plugins |
| configmap.k8s | object | `{"plugins":{"k8s":{"default-cpus":"100m","default-env-vars":[],"default-memory":"100Mi"}}}` | Kubernetes specific Flyte configuration |
| configmap.k8s.plugins.k8s | object | `{"default-cpus":"100m","default-env-vars":[],"default-memory":"100Mi"}` | Configuration section for all K8s specific plugins [Configuration structure](https://pkg.go.dev/github.com/lyft/flyteplugins/go/tasks/pluginmachinery/flytek8s/config) |
| configmap.remoteData.remoteData.region | string | `"us-east-1"` |  |
| configmap.remoteData.remoteData.scheme | string | `"local"` |  |
| configmap.remoteData.remoteData.signedUrls.durationMinutes | int | `3` |  |
| configmap.resource_manager | object | `{"propeller":{"resourcemanager":{"type":"noop"}}}` | Resource manager configuration |
| configmap.resource_manager.propeller | object | `{"resourcemanager":{"type":"noop"}}` | resource manager configuration |
| configmap.schedulerConfig.scheduler.metricsScope | string | `"flyte:"` |  |
| configmap.schedulerConfig.scheduler.profilerPort | int | `10254` |  |
| configmap.task_logs | object | `{"plugins":{"logs":{"cloudwatch-enabled":false,"kubernetes-enabled":false}}}` | Section that configures how the Task logs are displayed on the UI. This has to be changed based on your actual logging provider. Refer to [structure](https://pkg.go.dev/github.com/lyft/flyteplugins/go/tasks/logs#LogConfig) to understand how to configure various logging engines |
| configmap.task_logs.plugins.logs.cloudwatch-enabled | bool | `false` | One option is to enable cloudwatch logging for EKS, update the region and log group accordingly |
| configmap.task_resource_defaults | object | `{"task_resources":{"defaults":{"cpu":"100m","memory":"500Mi"},"limits":{"cpu":2,"gpu":1,"memory":"1Gi"}}}` | Task default resources configuration Refer to the full [structure](https://pkg.go.dev/github.com/lyft/flyteadmin@v0.3.37/pkg/runtime/interfaces#TaskResourceConfiguration). |
| configmap.task_resource_defaults.task_resources | object | `{"defaults":{"cpu":"100m","memory":"500Mi"},"limits":{"cpu":2,"gpu":1,"memory":"1Gi"}}` | Task default resources parameters |
| daskoperator | object | `{"enabled":false}` | Optional: Dask Plugin using the Dask Operator |
| daskoperator.enabled | bool | `false` | - enable or disable the dask operator deployment installation |
| databricks | object | `{"enabled":false,"plugin_config":{"plugins":{"databricks":{"databricksInstance":"dbc-a53b7a3c-614c","entrypointFile":"dbfs:///FileStore/tables/entrypoint.py"}}}}` | Optional: Databricks Plugin allows us to run the spark job on the Databricks platform. |
| datacatalog.additionalContainers | list | `[]` | Appends additional containers to the deployment spec. May include template values. |
| datacatalog.additionalVolumeMounts | list | `[]` | Appends additional volume mounts to the main container's spec. May include template values. |
| datacatalog.additionalVolumes | list | `[]` | Appends additional volumes to the deployment spec. May include template values. |
| datacatalog.affinity | object | `{}` | affinity for Datacatalog deployment |
| datacatalog.configPath | string | `"/etc/datacatalog/config/*.yaml"` | Default regex string for searching configuration files |
| datacatalog.enabled | bool | `true` |  |
| datacatalog.extraArgs | object | `{}` | Appends extra command line arguments to the main command |
| datacatalog.image.pullPolicy | string | `"IfNotPresent"` | Docker image pull policy |
| datacatalog.image.repository | string | `"cr.flyte.org/flyteorg/datacatalog"` | Docker image for Datacatalog deployment |
| datacatalog.image.tag | string | `"v1.13.2"` | Docker image tag |
| datacatalog.nodeSelector | object | `{}` | nodeSelector for Datacatalog deployment |
| datacatalog.podAnnotations | object | `{}` | Annotations for Datacatalog pods |
| datacatalog.podEnv | object | `{}` | Additional Datacatalog container environment variables |
| datacatalog.podLabels | object | `{}` | Labels for Datacatalog pods |
| datacatalog.priorityClassName | string | `""` | Sets priorityClassName for datacatalog pod(s). |
| datacatalog.replicaCount | int | `1` | Replicas count for Datacatalog deployment |
| datacatalog.resources | object | `{"limits":{"cpu":"500m","ephemeral-storage":"100Mi","memory":"500Mi"},"requests":{"cpu":"10m","ephemeral-storage":"50Mi","memory":"50Mi"}}` | Default resources requests and limits for Datacatalog deployment |
| datacatalog.securityContext | object | `{"fsGroup":1001,"fsGroupChangePolicy":"OnRootMismatch","runAsNonRoot":true,"runAsUser":1001,"seLinuxOptions":{"type":"spc_t"}}` | Sets securityContext for datacatalog pod(s). |
| datacatalog.service | object | `{"additionalPorts":[],"annotations":{"projectcontour.io/upstream-protocol.h2c":"grpc"},"type":"NodePort"}` | Service settings for Datacatalog |
| datacatalog.service.additionalPorts | list | `[]` | Appends additional ports to the service spec. |
| datacatalog.serviceAccount | object | `{"annotations":{},"create":true,"imagePullSecrets":[]}` | Configuration for service accounts for Datacatalog |
| datacatalog.serviceAccount.annotations | object | `{}` | Annotations for ServiceAccount attached to Datacatalog pods |
| datacatalog.serviceAccount.create | bool | `true` | Should a service account be created for Datacatalog |
| datacatalog.serviceAccount.imagePullSecrets | list | `[]` | ImagePullSecrets to automatically assign to the service account |
| datacatalog.tolerations | list | `[]` | tolerations for Datacatalog deployment |
| db.admin.database.dbname | string | `"flyteadmin"` |  |
| db.admin.database.host | string | `"postgres"` |  |
| db.admin.database.port | int | `5432` |  |
| db.admin.database.username | string | `"postgres"` |  |
| db.datacatalog.database.dbname | string | `"datacatalog"` |  |
| db.datacatalog.database.host | string | `"postgres"` |  |
| db.datacatalog.database.port | int | `5432` |  |
| db.datacatalog.database.username | string | `"postgres"` |  |
| deployRedoc | bool | `false` |  |
| external_events | object | `{"aws":{"region":"us-east-2"},"enable":false,"eventsPublisher":{"eventTypes":["all"],"topicName":"arn:aws:sns:us-east-2:123456:123-my-topic"},"type":"aws"}` | **Optional Component** External events are used to send events (unprocessed, as Admin see them) to an SNS topic (or gcp equivalent) The config is here as an example only - if not enabled, it won't be used. |
| flyteadmin.additionalContainers | list | `[]` | Appends additional containers to the deployment spec. May include template values. |
| flyteadmin.additionalVolumeMounts | list | `[]` | Appends additional volume mounts to the main container's spec. May include template values. |
| flyteadmin.additionalVolumes | list | `[]` | Appends additional volumes to the deployment spec. May include template values. |
| flyteadmin.affinity | object | `{}` | affinity for Flyteadmin deployment |
| flyteadmin.configPath | string | `"/etc/flyte/config/*.yaml"` | Default regex string for searching configuration files |
| flyteadmin.enabled | bool | `true` |  |
| flyteadmin.env | list | `[]` | Additional flyteadmin container environment variables  e.g. SendGrid's API key  - name: SENDGRID_API_KEY    value: "<your sendgrid api key>"  e.g. secret environment variable (you can combine it with .additionalVolumes): - name: SENDGRID_API_KEY   valueFrom:     secretKeyRef:       name: sendgrid-secret       key: api_key |
| flyteadmin.extraArgs | object | `{}` | Appends extra command line arguments to the serve command |
| flyteadmin.image.pullPolicy | string | `"IfNotPresent"` |  |
| flyteadmin.image.repository | string | `"cr.flyte.org/flyteorg/flyteadmin"` | Docker image for Flyteadmin deployment |
| flyteadmin.image.tag | string | `"v1.13.2"` |  |
| flyteadmin.initialProjects | list | `["flytesnacks","flytetester","flyteexamples"]` | Initial projects to create |
| flyteadmin.nodeSelector | object | `{}` | nodeSelector for Flyteadmin deployment |
| flyteadmin.podAnnotations | object | `{}` | Annotations for Flyteadmin pods |
| flyteadmin.podLabels | object | `{}` | Labels for Flyteadmin pods |
| flyteadmin.priorityClassName | string | `""` | Sets priorityClassName for flyteadmin pod(s). |
| flyteadmin.replicaCount | int | `1` | Replicas count for Flyteadmin deployment |
| flyteadmin.resources | object | `{"limits":{"cpu":"250m","ephemeral-storage":"100Mi","memory":"500Mi"},"requests":{"cpu":"10m","ephemeral-storage":"50Mi","memory":"50Mi"}}` | Default resources requests and limits for Flyteadmin deployment |
| flyteadmin.secrets | object | `{}` |  |
| flyteadmin.securityContext | object | `{"fsGroup":65534,"fsGroupChangePolicy":"Always","runAsNonRoot":true,"runAsUser":1001,"seLinuxOptions":{"type":"spc_t"}}` | Sets securityContext for flyteadmin pod(s). |
| flyteadmin.service | object | `{"additionalPorts":[],"annotations":{"projectcontour.io/upstream-protocol.h2c":"grpc"},"loadBalancerSourceRanges":[],"type":"ClusterIP"}` | Service settings for Flyteadmin |
| flyteadmin.service.additionalPorts | list | `[]` | Appends additional ports to the service spec. |
| flyteadmin.serviceAccount | object | `{"alwaysCreate":false,"annotations":{},"clusterRole":{"apiGroups":["","flyte.lyft.com","rbac.authorization.k8s.io"],"resources":["configmaps","flyteworkflows","namespaces","pods","resourcequotas","roles","rolebindings","secrets","services","serviceaccounts","spark-role","limitranges"],"verbs":["*"]},"create":true,"createClusterRole":true,"imagePullSecrets":[]}` | Configuration for service accounts for FlyteAdmin |
| flyteadmin.serviceAccount.alwaysCreate | bool | `false` | Should a service account always be created for flyteadmin even without an actual flyteadmin deployment running (e.g. for multi-cluster setups) |
| flyteadmin.serviceAccount.annotations | object | `{}` | Annotations for ServiceAccount attached to Flyteadmin pods |
| flyteadmin.serviceAccount.clusterRole | object | `{"apiGroups":["","flyte.lyft.com","rbac.authorization.k8s.io"],"resources":["configmaps","flyteworkflows","namespaces","pods","resourcequotas","roles","rolebindings","secrets","services","serviceaccounts","spark-role","limitranges"],"verbs":["*"]}` | Configuration for ClusterRole created for Flyteadmin |
| flyteadmin.serviceAccount.clusterRole.apiGroups | list | `["","flyte.lyft.com","rbac.authorization.k8s.io"]` | Specifies the API groups that this ClusterRole can access |
| flyteadmin.serviceAccount.clusterRole.resources | list | `["configmaps","flyteworkflows","namespaces","pods","resourcequotas","roles","rolebindings","secrets","services","serviceaccounts","spark-role","limitranges"]` | Specifies the resources that this ClusterRole can access |
| flyteadmin.serviceAccount.clusterRole.verbs | list | `["*"]` | Specifies the verbs (actions) that this ClusterRole can perform on the specified resources |
| flyteadmin.serviceAccount.create | bool | `true` | Should a service account be created for flyteadmin |
| flyteadmin.serviceAccount.createClusterRole | bool | `true` | Should a ClusterRole be created for Flyteadmin |
| flyteadmin.serviceAccount.imagePullSecrets | list | `[]` | ImagePullSecrets to automatically assign to the service account |
| flyteadmin.serviceMonitor | object | `{"enabled":false,"interval":"60s","labels":{},"scrapeTimeout":"30s"}` | Settings for flyteadmin service monitor |
| flyteadmin.serviceMonitor.enabled | bool | `false` | If enabled create the flyteadmin service monitor |
| flyteadmin.serviceMonitor.interval | string | `"60s"` | Sets the interval at which metrics will be scraped by prometheus |
| flyteadmin.serviceMonitor.labels | object | `{}` | Sets the labels for the service monitor which are required by the prometheus to auto-detect the service monitor and start scrapping the metrics |
| flyteadmin.serviceMonitor.scrapeTimeout | string | `"30s"` | Sets the timeout after which request to scrape metrics will time out |
| flyteadmin.tolerations | list | `[]` | tolerations for Flyteadmin deployment |
| flyteagent.enabled | bool | `false` |  |
| flyteagent.plugin_config.plugins.agent-service | object | `{"defaultAgent":{"endpoint":"dns:///flyteagent.flyte.svc.cluster.local:8000","insecure":true},"supportedTaskTypes":[]}` | Agent service configuration for propeller. |
| flyteagent.plugin_config.plugins.agent-service.defaultAgent | object | `{"endpoint":"dns:///flyteagent.flyte.svc.cluster.local:8000","insecure":true}` | The default agent service to use for plugin tasks. |
| flyteagent.plugin_config.plugins.agent-service.defaultAgent.endpoint | string | `"dns:///flyteagent.flyte.svc.cluster.local:8000"` | The agent service endpoint propeller should connect to. |
| flyteagent.plugin_config.plugins.agent-service.defaultAgent.insecure | bool | `true` | Whether the connection from propeller to the agent service should use TLS. |
| flyteagent.plugin_config.plugins.agent-service.supportedTaskTypes | list | `[]` | The task types supported by the default agent. As of #5460 these are discovered automatically and don't need to be configured. |
| flyteagent.podLabels | object | `{}` | Labels for flyteagent pods |
| flyteconsole.affinity | object | `{}` | affinity for Flyteconsole deployment |
| flyteconsole.enabled | bool | `true` |  |
| flyteconsole.ga.enabled | bool | `false` |  |
| flyteconsole.ga.tracking_id | string | `"G-0QW4DJWJ20"` |  |
| flyteconsole.image.pullPolicy | string | `"IfNotPresent"` |  |
| flyteconsole.image.repository | string | `"cr.flyte.org/flyteorg/flyteconsole"` | Docker image for Flyteconsole deployment |
| flyteconsole.image.tag | string | `"v1.17.1"` |  |
| flyteconsole.imagePullSecrets | list | `[]` | ImagePullSecrets to assign to the Flyteconsole deployment |
| flyteconsole.livenessProbe | object | `{}` |  |
| flyteconsole.nodeSelector | object | `{}` | nodeSelector for Flyteconsole deployment |
| flyteconsole.podAnnotations | object | `{}` | Annotations for Flyteconsole pods |
| flyteconsole.podEnv | object | `{}` | Additional Flyteconsole container environment variables |
| flyteconsole.podLabels | object | `{}` | Labels for Flyteconsole pods |
| flyteconsole.priorityClassName | string | `""` | Sets priorityClassName for flyte console pod(s). |
| flyteconsole.readinessProbe | object | `{}` |  |
| flyteconsole.replicaCount | int | `1` | Replicas count for Flyteconsole deployment |
| flyteconsole.resources | object | `{"limits":{"cpu":"500m","memory":"250Mi"},"requests":{"cpu":"10m","memory":"50Mi"}}` | Default resources requests and limits for Flyteconsole deployment |
| flyteconsole.securityContext | object | `{"fsGroupChangePolicy":"OnRootMismatch","runAsNonRoot":true,"runAsUser":1000,"seLinuxOptions":{"type":"spc_t"}}` | Sets securityContext for flyteconsole pod(s). |
| flyteconsole.service | object | `{"annotations":{},"type":"ClusterIP"}` | Service settings for Flyteconsole |
| flyteconsole.serviceMonitor | object | `{"enabled":false,"interval":"60s","labels":{},"scrapeTimeout":"30s"}` | Settings for flyteconsole service monitor |
| flyteconsole.serviceMonitor.enabled | bool | `false` | If enabled create the flyteconsole service monitor |
| flyteconsole.serviceMonitor.interval | string | `"60s"` | Sets the interval at which metrics will be scraped by prometheus |
| flyteconsole.serviceMonitor.labels | object | `{}` | Sets the labels for the service monitor which are required by the prometheus to auto-detect the service monitor and start scrapping the metrics |
| flyteconsole.serviceMonitor.scrapeTimeout | string | `"30s"` | Sets the timeout after which request to scrape metrics will time out |
| flyteconsole.tolerations | list | `[]` | tolerations for Flyteconsole deployment |
| flytepropeller.additionalContainers | list | `[]` | Appends additional containers to the deployment spec. May include template values. |
| flytepropeller.additionalVolumeMounts | list | `[]` | Appends additional volume mounts to the main container's spec. May include template values. |
| flytepropeller.additionalVolumes | list | `[]` | Appends additional volumes to the deployment spec. May include template values. |
| flytepropeller.affinity | object | `{}` | affinity for Flytepropeller deployment |
| flytepropeller.clusterName | string | `""` | Defines the cluster name used in events sent to Admin |
| flytepropeller.configPath | string | `"/etc/flyte/config/*.yaml"` | Default regex string for searching configuration files |
| flytepropeller.createCRDs | bool | `true` | Whether to install the flyteworkflows CRD with helm |
| flytepropeller.enabled | bool | `true` |  |
| flytepropeller.extraArgs | object | `{}` | Appends extra command line arguments to the main command |
| flytepropeller.image.pullPolicy | string | `"IfNotPresent"` |  |
| flytepropeller.image.repository | string | `"cr.flyte.org/flyteorg/flytepropeller"` | Docker image for Flytepropeller deployment |
| flytepropeller.image.tag | string | `"v1.13.2"` |  |
| flytepropeller.manager | bool | `false` |  |
| flytepropeller.nodeSelector | object | `{}` | nodeSelector for Flytepropeller deployment |
| flytepropeller.podAnnotations | object | `{}` | Annotations for Flytepropeller pods |
| flytepropeller.podEnv | object | `{}` | Additional Flytepropeller container environment variables |
| flytepropeller.podLabels | object | `{}` | Labels for Flytepropeller pods |
| flytepropeller.priorityClassName | string | `""` | Sets priorityClassName for propeller pod(s). |
| flytepropeller.prometheus.enabled | bool | `false` |  |
| flytepropeller.replicaCount | int | `1` | Replicas count for Flytepropeller deployment |
| flytepropeller.resources | object | `{"limits":{"cpu":"200m","ephemeral-storage":"100Mi","memory":"200Mi"},"requests":{"cpu":"10m","ephemeral-storage":"50Mi","memory":"100Mi"}}` | Default resources requests and limits for Flytepropeller deployment |
| flytepropeller.securityContext | object | `{"fsGroup":65534,"fsGroupChangePolicy":"Always","runAsUser":1001}` | Sets securityContext for flytepropeller pod(s). |
| flytepropeller.service | object | `{"additionalPorts":[],"enabled":false}` | Settings for flytepropeller service |
| flytepropeller.service.additionalPorts | list | `[]` | Appends additional ports to the service spec. |
| flytepropeller.service.enabled | bool | `false` | If enabled create the flytepropeller service |
| flytepropeller.serviceAccount | object | `{"annotations":{},"create":true,"imagePullSecrets":[]}` | Configuration for service accounts for FlytePropeller |
| flytepropeller.serviceAccount.annotations | object | `{}` | Annotations for ServiceAccount attached to FlytePropeller pods |
| flytepropeller.serviceAccount.create | bool | `true` | Should a service account be created for FlytePropeller |
| flytepropeller.serviceAccount.imagePullSecrets | list | `[]` | ImagePullSecrets to automatically assign to the service account |
| flytepropeller.serviceMonitor | object | `{"enabled":false,"interval":"60s","labels":{},"scrapeTimeout":"30s"}` | Settings for flytepropeller service monitor |
| flytepropeller.serviceMonitor.enabled | bool | `false` | If enabled create the flyetepropeller service monitor |
| flytepropeller.serviceMonitor.interval | string | `"60s"` | Sets the interval at which metrics will be scraped by prometheus |
| flytepropeller.serviceMonitor.labels | object | `{}` | Sets the labels for the service monitor which are required by the prometheus to auto-detect the service monitor and start scrapping the metrics |
| flytepropeller.serviceMonitor.scrapeTimeout | string | `"30s"` | Sets the timeout after which request to scrape metrics will time out |
| flytepropeller.terminationMessagePolicy | string | `"FallbackToLogsOnError"` | Error reporting |
| flytepropeller.tolerations | list | `[]` | tolerations for Flytepropeller deployment |
| flytescheduler.additionalContainers | list | `[]` | Appends additional containers to the deployment spec. May include template values. |
| flytescheduler.additionalVolumeMounts | list | `[]` | Appends additional volume mounts to the main container's spec. May include template values. |
| flytescheduler.additionalVolumes | list | `[]` | Appends additional volumes to the deployment spec. May include template values. |
| flytescheduler.affinity | object | `{}` | affinity for Flytescheduler deployment |
| flytescheduler.configPath | string | `"/etc/flyte/config/*.yaml"` | Default regex string for searching configuration files |
| flytescheduler.image.pullPolicy | string | `"IfNotPresent"` | Docker image pull policy |
| flytescheduler.image.repository | string | `"cr.flyte.org/flyteorg/flytescheduler"` | Docker image for Flytescheduler deployment |
| flytescheduler.image.tag | string | `"v1.13.2"` | Docker image tag |
| flytescheduler.nodeSelector | object | `{}` | nodeSelector for Flytescheduler deployment |
| flytescheduler.podAnnotations | object | `{}` | Annotations for Flytescheduler pods |
| flytescheduler.podEnv | object | `{}` | Additional Flytescheduler container environment variables |
| flytescheduler.podLabels | object | `{}` | Labels for Flytescheduler pods |
| flytescheduler.priorityClassName | string | `""` | Sets priorityClassName for flyte scheduler pod(s). |
| flytescheduler.resources | object | `{"limits":{"cpu":"250m","ephemeral-storage":"100Mi","memory":"500Mi"},"requests":{"cpu":"10m","ephemeral-storage":"50Mi","memory":"50Mi"}}` | Default resources requests and limits for Flytescheduler deployment |
| flytescheduler.runPrecheck | bool | `true` | Whether to inject an init container which waits on flyteadmin |
| flytescheduler.secrets | object | `{}` |  |
| flytescheduler.securityContext | object | `{"fsGroup":65534,"fsGroupChangePolicy":"Always","runAsNonRoot":true,"runAsUser":1001,"seLinuxOptions":{"type":"spc_t"}}` | Sets securityContext for flytescheduler pod(s). |
| flytescheduler.serviceAccount | object | `{"annotations":{},"create":true,"imagePullSecrets":[]}` | Configuration for service accounts for Flytescheduler |
| flytescheduler.serviceAccount.annotations | object | `{}` | Annotations for ServiceAccount attached to Flytescheduler pods |
| flytescheduler.serviceAccount.create | bool | `true` | Should a service account be created for Flytescheduler |
| flytescheduler.serviceAccount.imagePullSecrets | list | `[]` | ImagePullSecrets to automatically assign to the service account |
| flytescheduler.tolerations | list | `[]` | tolerations for Flytescheduler deployment |
| secrets.adminOauthClientCredentials.clientId | string | `"flytepropeller"` |  |
| secrets.adminOauthClientCredentials.clientSecret | string | `"foobar"` |  |
| secrets.adminOauthClientCredentials.enabled | bool | `true` |  |
| sparkoperator | object | `{"enabled":false,"plugin_config":{"plugins":{"spark":{"spark-config-default":[{"spark.hadoop.fs.s3a.aws.credentials.provider":"com.amazonaws.auth.DefaultAWSCredentialsProviderChain"},{"spark.hadoop.mapreduce.fileoutputcommitter.algorithm.version":"2"},{"spark.kubernetes.allocation.batch.size":"50"},{"spark.hadoop.fs.s3a.acl.default":"BucketOwnerFullControl"},{"spark.hadoop.fs.s3n.impl":"org.apache.hadoop.fs.s3a.S3AFileSystem"},{"spark.hadoop.fs.AbstractFileSystem.s3n.impl":"org.apache.hadoop.fs.s3a.S3A"},{"spark.hadoop.fs.s3.impl":"org.apache.hadoop.fs.s3a.S3AFileSystem"},{"spark.hadoop.fs.AbstractFileSystem.s3.impl":"org.apache.hadoop.fs.s3a.S3A"},{"spark.hadoop.fs.s3a.impl":"org.apache.hadoop.fs.s3a.S3AFileSystem"},{"spark.hadoop.fs.AbstractFileSystem.s3a.impl":"org.apache.hadoop.fs.s3a.S3A"},{"spark.hadoop.fs.s3a.multipart.threshold":"536870912"},{"spark.blacklist.enabled":"true"},{"spark.blacklist.timeout":"5m"},{"spark.task.maxfailures":"8"}]}}}}` | Optional: Spark Plugin using the Spark Operator |
| sparkoperator.enabled | bool | `false` | - enable or disable Sparkoperator deployment installation |
| sparkoperator.plugin_config | object | `{"plugins":{"spark":{"spark-config-default":[{"spark.hadoop.fs.s3a.aws.credentials.provider":"com.amazonaws.auth.DefaultAWSCredentialsProviderChain"},{"spark.hadoop.mapreduce.fileoutputcommitter.algorithm.version":"2"},{"spark.kubernetes.allocation.batch.size":"50"},{"spark.hadoop.fs.s3a.acl.default":"BucketOwnerFullControl"},{"spark.hadoop.fs.s3n.impl":"org.apache.hadoop.fs.s3a.S3AFileSystem"},{"spark.hadoop.fs.AbstractFileSystem.s3n.impl":"org.apache.hadoop.fs.s3a.S3A"},{"spark.hadoop.fs.s3.impl":"org.apache.hadoop.fs.s3a.S3AFileSystem"},{"spark.hadoop.fs.AbstractFileSystem.s3.impl":"org.apache.hadoop.fs.s3a.S3A"},{"spark.hadoop.fs.s3a.impl":"org.apache.hadoop.fs.s3a.S3AFileSystem"},{"spark.hadoop.fs.AbstractFileSystem.s3a.impl":"org.apache.hadoop.fs.s3a.S3A"},{"spark.hadoop.fs.s3a.multipart.threshold":"536870912"},{"spark.blacklist.enabled":"true"},{"spark.blacklist.timeout":"5m"},{"spark.task.maxfailures":"8"}]}}}` | Spark plugin configuration |
| sparkoperator.plugin_config.plugins.spark.spark-config-default | list | `[{"spark.hadoop.fs.s3a.aws.credentials.provider":"com.amazonaws.auth.DefaultAWSCredentialsProviderChain"},{"spark.hadoop.mapreduce.fileoutputcommitter.algorithm.version":"2"},{"spark.kubernetes.allocation.batch.size":"50"},{"spark.hadoop.fs.s3a.acl.default":"BucketOwnerFullControl"},{"spark.hadoop.fs.s3n.impl":"org.apache.hadoop.fs.s3a.S3AFileSystem"},{"spark.hadoop.fs.AbstractFileSystem.s3n.impl":"org.apache.hadoop.fs.s3a.S3A"},{"spark.hadoop.fs.s3.impl":"org.apache.hadoop.fs.s3a.S3AFileSystem"},{"spark.hadoop.fs.AbstractFileSystem.s3.impl":"org.apache.hadoop.fs.s3a.S3A"},{"spark.hadoop.fs.s3a.impl":"org.apache.hadoop.fs.s3a.S3AFileSystem"},{"spark.hadoop.fs.AbstractFileSystem.s3a.impl":"org.apache.hadoop.fs.s3a.S3A"},{"spark.hadoop.fs.s3a.multipart.threshold":"536870912"},{"spark.blacklist.enabled":"true"},{"spark.blacklist.timeout":"5m"},{"spark.task.maxfailures":"8"}]` | Spark default configuration |
| storage | object | `{"bucketName":"my-s3-bucket","cache":{"maxSizeMBs":0,"targetGCPercent":70},"custom":{},"enableMultiContainer":false,"gcs":null,"limits":{"maxDownloadMBs":10},"s3":{"accessKey":"","authType":"iam","region":"us-east-1","secretKey":""},"type":"sandbox"}` | ----------------------------------------------------  STORAGE SETTINGS  |
| storage.bucketName | string | `"my-s3-bucket"` | bucketName defines the storage bucket flyte will use. Required for all types except for sandbox. |
| storage.custom | object | `{}` | Settings for storage type custom. See https://github.com/graymeta/stow for supported storage providers/settings. |
| storage.enableMultiContainer | bool | `false` | toggles multi-container storage config |
| storage.gcs | string | `nil` | settings for storage type gcs |
| storage.limits | object | `{"maxDownloadMBs":10}` | default limits being applied to storage config |
| storage.s3 | object | `{"accessKey":"","authType":"iam","region":"us-east-1","secretKey":""}` | settings for storage type s3 |
| storage.s3.accessKey | string | `""` | AWS IAM user access key ID to use for S3 bucket auth, only used if authType is set to accesskey |
| storage.s3.authType | string | `"iam"` | type of authentication to use for S3 buckets, can either be iam or accesskey |
| storage.s3.secretKey | string | `""` | AWS IAM user secret access key to use for S3 bucket auth, only used if authType is set to accesskey |
| storage.type | string | `"sandbox"` | Sets the storage type. Supported values are sandbox, s3, gcs and custom. |
| webhook.autoscaling.enabled | bool | `false` |  |
| webhook.autoscaling.maxReplicas | int | `10` |  |
| webhook.autoscaling.metrics[0].resource.name | string | `"cpu"` |  |
| webhook.autoscaling.metrics[0].resource.target.averageUtilization | int | `80` |  |
| webhook.autoscaling.metrics[0].resource.target.type | string | `"Utilization"` |  |
| webhook.autoscaling.metrics[0].type | string | `"Resource"` |  |
| webhook.autoscaling.metrics[1].resource.name | string | `"memory"` |  |
| webhook.autoscaling.metrics[1].resource.target.averageUtilization | int | `80` |  |
| webhook.autoscaling.metrics[1].resource.target.type | string | `"Utilization"` |  |
| webhook.autoscaling.metrics[1].type | string | `"Resource"` |  |
| webhook.autoscaling.minReplicas | int | `1` |  |
| webhook.enabled | bool | `true` | enable or disable secrets webhook |
| webhook.nodeSelector | object | `{}` | nodeSelector for webhook deployment |
| webhook.podAnnotations | object | `{}` | Annotations for webhook pods |
| webhook.podEnv | object | `{}` | Additional webhook container environment variables |
| webhook.podLabels | object | `{}` | Labels for webhook pods |
| webhook.priorityClassName | string | `""` | Sets priorityClassName for webhook pod |
| webhook.prometheus.enabled | bool | `false` |  |
| webhook.resources.requests.cpu | string | `"200m"` |  |
| webhook.resources.requests.ephemeral-storage | string | `"500Mi"` |  |
| webhook.resources.requests.memory | string | `"500Mi"` |  |
| webhook.securityContext | object | `{"fsGroup":65534,"fsGroupChangePolicy":"Always","runAsNonRoot":true,"runAsUser":1001,"seLinuxOptions":{"type":"spc_t"}}` | Sets securityContext for webhook pod(s). |
| webhook.service | object | `{"annotations":{"projectcontour.io/upstream-protocol.h2c":"grpc"},"type":"ClusterIP"}` | Service settings for the webhook |
| webhook.serviceAccount | object | `{"annotations":{},"create":true,"imagePullSecrets":[]}` | Configuration for service accounts for the webhook |
| webhook.serviceAccount.annotations | object | `{}` | Annotations for ServiceAccount attached to the webhook |
| webhook.serviceAccount.create | bool | `true` | Should a service account be created for the webhook |
| webhook.serviceAccount.imagePullSecrets | list | `[]` | ImagePullSecrets to automatically assign to the service account |
| workflow_notifications | object | `{"config":{},"enabled":false}` | **Optional Component** Workflow notifications module is an optional dependency. Flyte uses cloud native pub-sub systems to notify users of various events in their workflows |
| workflow_scheduler | object | `{"config":{},"enabled":false,"type":""}` | **Optional Component** Flyte uses a cloud hosted Cron scheduler to run workflows on a schedule. The following module is optional. Without, this module, you will not have scheduled launchplans / workflows. Docs: https://docs.flyte.org/en/latest/howto/enable_and_use_schedules.html#setting-up-scheduled-workflows |
