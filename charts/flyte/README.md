# flyte

![Version: v0.1.10](https://img.shields.io/badge/Version-v0.1.10-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)

A Helm chart for Flyte Sandbox

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../flyte-core | flyte(flyte-core) | v0.1.10 |
| https://charts.bitnami.com/bitnami | contour | 4.1.2 |
| https://googlecloudplatform.github.io/spark-on-k8s-operator | sparkoperator(spark-operator) | 1.0.6 |
| https://kubernetes.github.io/dashboard/ | kubernetes-dashboard | 4.0.2 |

### SANDBOX INSTALLATION:
- [Install helm 3](https://helm.sh/docs/intro/install/)
- Install Flyte sandbox:

```bash
helm repo add flyte https://flyteorg.github.io/flyte
helm install -n flyte -f values.yaml --create-namespace flyte flyte/flyte
```

Customize your installation by changing settings in a new file `values-sandbox.yaml`.
You can use the helm diff plugin to review any value changes you've made to your values:

```bash
helm plugin install https://github.com/databus23/helm-diff
helm diff upgrade -f values-sandbox.yaml flyte .
```

Then apply your changes:
```bash
helm upgrade -f values-sandbox.yaml flyte .
```

#### Alternative: Generate raw kubernetes yaml with helm template
- `helm template --name-template=flyte-sandbox . -n flyte -f values-sandbox.yaml > flyte_generated_sandbox.yaml`
- Deploy the manifest `kubectl apply -f flyte_generated_sandbox.yaml`

- When all pods are running - run end2end tests: `kubectl apply -f ../end2end/tests/endtoend.yaml`
- If running on minikube, get flyte host using `minikube service contour -n heptio-contour --url`. And then visit `http://<HOST>/console`

### CONFIGURATION NOTES:
- The docker images, their tags and other default parameters are configured in `values.yaml` file.
- Each Flyte installation type should have separate `values-*.yaml` file: for sandbox, EKS and etc. The configuration in `values.yaml` and the choosen config `values-*.yaml` are merged when generating the deployment manifest.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| contour.affinity | object | `{}` | affinity for Contour deployment |
| contour.contour.resources | object | `{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"10m","memory":"50Mi"}}` | Default resources requests and limits for Contour |
| contour.contour.resources.limits | object | `{"cpu":"100m","memory":"100Mi"}` | Limits are the maximum set of resources needed for this pod |
| contour.contour.resources.requests | object | `{"cpu":"10m","memory":"50Mi"}` | Requests are the minimum set of resources needed for this pod |
| contour.enabled | bool | `true` | - enable or disable Contour deployment installation |
| contour.envoy.resources | object | `{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"10m","memory":"50Mi"}}` | Default resources requests and limits for Envoy |
| contour.envoy.resources.limits | object | `{"cpu":"100m","memory":"100Mi"}` | Limits are the maximum set of resources needed for this pod |
| contour.envoy.resources.requests | object | `{"cpu":"10m","memory":"50Mi"}` | Requests are the minimum set of resources needed for this pod |
| contour.envoy.service.nodePorts.http | int | `30081` |  |
| contour.envoy.service.ports.http | int | `80` |  |
| contour.envoy.service.type | string | `"NodePort"` |  |
| contour.nodeSelector | object | `{}` | nodeSelector for Contour deployment |
| contour.podAnnotations | object | `{}` | Annotations for Contour pods |
| contour.replicaCount | int | `1` | Replicas count for Contour deployment |
| contour.serviceAccountAnnotations | object | `{}` | Annotations for ServiceAccount attached to Contour pods |
| contour.tolerations | list | `[]` | tolerations for Contour deployment |
| flyte | object | `{"cluster_resource_manager":{"config":{"cluster_resources":{"customData":[{"production":[{"projectQuotaCpu":{"value":"5"}},{"projectQuotaMemory":{"value":"4000Mi"}}]},{"staging":[{"projectQuotaCpu":{"value":"2"}},{"projectQuotaMemory":{"value":"3000Mi"}}]},{"development":[{"projectQuotaCpu":{"value":"4"}},{"projectQuotaMemory":{"value":"3000Mi"}}]}],"refresh":"5m","refreshInterval":"5m","templatePath":"/etc/flyte/clusterresource/templates"}},"enabled":true,"templates":[{"key":"aa_namespace","value":"apiVersion: v1\nkind: Namespace\nmetadata:\n  name: {{ namespace }}\nspec:\n  finalizers:\n  - kubernetes\n"},{"key":"ab_project_resource_quota","value":"apiVersion: v1\nkind: ResourceQuota\nmetadata:\n  name: project-quota\n  namespace: {{ namespace }}\nspec:\n  hard:\n    limits.cpu: {{ projectQuotaCpu }}\n    limits.memory: {{ projectQuotaMemory }}\n"}]},"common":{"databaseSecret":{"name":"","secretManifest":{}},"flyteNamespaceTemplate":{"enabled":false},"ingress":{"albSSLRedirect":false,"annotations":{"nginx.ingress.kubernetes.io/app-root":"/console"},"enabled":true,"host":"","separateGrpcIngress":false,"separateGrpcIngressAnnotations":{"nginx.ingress.kubernetes.io/backend-protocol":"GRPC"},"tls":{"enabled":false},"webpackHMR":true}},"configmap":{"adminServer":{"auth":{"appAuth":{"thirdPartyConfig":{"flyteClient":{"clientId":"flytectl","redirectUri":"http://localhost:53593/callback","scopes":["offline","all"]}}},"authorizedUris":["https://localhost:30081","http://flyteadmin:80","http://flyteadmin.flyte.svc.cluster.local:80"],"userAuth":{"openId":{"baseUrl":"https://accounts.google.com","clientId":"657465813211-6eog7ek7li5k7i7fvgv2921075063hpe.apps.googleusercontent.com","scopes":["profile","openid"]}}},"flyteadmin":{"eventVersion":2,"metadataStoragePrefix":["metadata","admin"],"metricsScope":"flyte:","profilerPort":10254,"roleNameKey":"iam.amazonaws.com/role","testing":{"host":"http://flyteadmin"}},"server":{"grpcPort":8089,"httpPort":8088,"security":{"allowCors":true,"allowedHeaders":["Content-Type","flyte-authorization"],"allowedOrigins":["*"],"secure":false,"useAuth":false}}},"catalog":{"catalog-cache":{"endpoint":"datacatalog:89","insecure":true,"type":"datacatalog"}},"console":{"BASE_URL":"/console","CONFIG_DIR":"/etc/flyte/config","DISABLE_AUTH":"1"},"copilot":{"plugins":{"k8s":{"co-pilot":{"image":"cr.flyte.org/flyteorg/flytecopilot:v0.0.24","name":"flyte-copilot-","start-timeout":"30s"}}}},"core":{"propeller":{"downstream-eval-duration":"30s","enable-admin-launcher":true,"leader-election":{"enabled":true,"lease-duration":"15s","lock-config-map":{"name":"propeller-leader","namespace":"flyte"},"renew-deadline":"10s","retry-period":"2s"},"limit-namespace":"all","max-workflow-retries":30,"metadata-prefix":"metadata/propeller","metrics-prefix":"flyte","prof-port":10254,"queue":{"batch-size":-1,"batching-interval":"2s","queue":{"base-delay":"5s","capacity":1000,"max-delay":"120s","rate":100,"type":"maxof"},"sub-queue":{"capacity":100,"rate":10,"type":"bucket"},"type":"batch"},"rawoutput-prefix":"s3://my-s3-bucket/","workers":4,"workflow-reeval-duration":"30s"},"webhook":{"certDir":"/etc/webhook/certs","serviceName":"flyte-pod-webhook"}},"datacatalogServer":{"application":{"grpcPort":8089,"grpcServerReflection":true,"httpPort":8080},"datacatalog":{"metrics-scope":"datacatalog","profiler-port":10254,"storage-prefix":"metadata/datacatalog"}},"domain":{"domains":[{"id":"development","name":"development"},{"id":"staging","name":"staging"},{"id":"production","name":"production"}]},"enabled_plugins":{"tasks":{"task-plugins":{"default-for-task-types":{"container":"container","container_array":"k8s-array","sidecar":"sidecar"},"enabled-plugins":["container","sidecar","k8s-array"]}}},"k8s":{"plugins":{"k8s":{"default-cpus":"100m","default-env-vars":[{"FLYTE_AWS_ENDPOINT":"http://minio.flyte:9000"},{"FLYTE_AWS_ACCESS_KEY_ID":"minio"},{"FLYTE_AWS_SECRET_ACCESS_KEY":"miniostorage"}],"default-memory":"200Mi"}}},"logger":{"logger":{"level":4,"show-source":true}},"remoteData":{"remoteData":{"region":"us-east-1","scheme":"local","signedUrls":{"durationMinutes":3}}},"resource_manager":{"propeller":{"resourcemanager":{"redis":null,"type":"noop"}}},"task_logs":{"plugins":{"logs":{"cloudwatch-enabled":false,"kubernetes-enabled":true,"kubernetes-template-uri":"http://localhost:30082/#/log/{{ \"{{\" }} .namespace {{ \"}}\" }}/{{ \"{{\" }} .podName {{ \"}}\" }}/pod?namespace={{ \"{{\" }} .namespace {{ \"}}\" }}"}}},"task_resource_defaults":{"task_resources":{"defaults":{"cpu":"100m","memory":"200Mi","storage":"5Mi"},"limits":{"cpu":2,"gpu":1,"memory":"1Gi","storage":"20Mi"}}}},"datacatalog":{"affinity":{},"configPath":"/etc/datacatalog/config/*.yaml","image":{"pullPolicy":"IfNotPresent","repository":"cr.flyte.org/flyteorg/datacatalog","tag":"v0.3.19"},"nodeSelector":{},"podAnnotations":{},"replicaCount":1,"resources":{"limits":{"cpu":"500m","ephemeral-storage":"100Mi","memory":"500Mi"},"requests":{"cpu":"10m","ephemeral-storage":"50Mi","memory":"50Mi"}},"service":{"annotations":{"projectcontour.io/upstream-protocol.h2c":"grpc"},"type":"NodePort"},"serviceAccount":{"annotations":{},"create":true,"imagePullSecrets":{}},"tolerations":[]},"db":{"admin":{"database":{"dbname":"flyteadmin","host":"postgres","port":5432,"username":"postgres"}},"datacatalog":{"database":{"dbname":"datacatalog","host":"postgres","port":5432,"username":"postgres"}}},"flyteadmin":{"additionalVolumeMounts":[],"additionalVolumes":[],"affinity":{},"configPath":"/etc/flyte/config/*.yaml","deployRedoc":true,"image":{"pullPolicy":"IfNotPresent","repository":"cr.flyte.org/flyteorg/flyteadmin","tag":"v0.6.89"},"initialProjects":["flytesnacks","flytetester","flyteexamples"],"nodeSelector":{},"podAnnotations":{},"replicaCount":1,"resources":{"limits":{"cpu":"250m","ephemeral-storage":"100Mi","memory":"500Mi"},"requests":{"cpu":"10m","ephemeral-storage":"50Mi","memory":"50Mi"}},"secrets":{},"service":{"annotations":{"projectcontour.io/upstream-protocol.h2c":"grpc"},"loadBalancerSourceRanges":[],"type":"ClusterIP"},"serviceAccount":{"annotations":{},"create":true,"imagePullSecrets":{}},"tolerations":[]},"flyteconsole":{"affinity":{},"ga":{"enabled":true,"tracking_id":"G-0QW4DJWJ20"},"image":{"pullPolicy":"IfNotPresent","repository":"cr.flyte.org/flyteorg/flyteconsole","tag":"v0.40.0"},"nodeSelector":{},"podAnnotations":{},"replicaCount":1,"resources":{"limits":{"cpu":"500m","memory":"250Mi"},"requests":{"cpu":"10m","memory":"50Mi"}},"service":{"annotations":{},"type":"ClusterIP"},"tolerations":[]},"flytepropeller":{"affinity":{},"cacheSizeMbs":0,"configPath":"/etc/flyte/config/*.yaml","image":{"pullPolicy":"IfNotPresent","repository":"cr.flyte.org/flyteorg/flytepropeller","tag":"v0.16.14"},"manager":false,"nodeSelector":{},"podAnnotations":{},"replicaCount":1,"resources":{"limits":{"cpu":"200m","ephemeral-storage":"100Mi","memory":"200Mi"},"requests":{"cpu":"10m","ephemeral-storage":"50Mi","memory":"50Mi"}},"serviceAccount":{"annotations":{},"create":true,"imagePullSecrets":{}},"tolerations":[]},"flytescheduler":{"affinity":{},"configPath":"/etc/flyte/config/*.yaml","image":{"pullPolicy":"IfNotPresent","repository":"cr.flyte.org/flyteorg/flytescheduler","tag":"v0.6.89"},"nodeSelector":{},"podAnnotations":{},"resources":{"limits":{"cpu":"250m","ephemeral-storage":"100Mi","memory":"500Mi"},"requests":{"cpu":"10m","ephemeral-storage":"50Mi","memory":"50Mi"}},"secrets":{},"serviceAccount":{"annotations":{},"create":true,"imagePullSecrets":{}},"tolerations":[]},"storage":{"bucketName":"my-s3-bucket","custom":{},"gcs":null,"s3":{"region":"us-east-1"},"type":"sandbox"},"webhook":{"enabled":true,"service":{"annotations":{"projectcontour.io/upstream-protocol.h2c":"grpc"},"type":"ClusterIP"},"serviceAccount":{"annotations":{},"create":true,"imagePullSecrets":{}}},"workflow_notifications":{"config":{},"enabled":false},"workflow_scheduler":{"enabled":true,"type":"native"}}` | ------------------------------------------------------------------- |
| flyte.cluster_resource_manager | object | `{"config":{"cluster_resources":{"customData":[{"production":[{"projectQuotaCpu":{"value":"5"}},{"projectQuotaMemory":{"value":"4000Mi"}}]},{"staging":[{"projectQuotaCpu":{"value":"2"}},{"projectQuotaMemory":{"value":"3000Mi"}}]},{"development":[{"projectQuotaCpu":{"value":"4"}},{"projectQuotaMemory":{"value":"3000Mi"}}]}],"refresh":"5m","refreshInterval":"5m","templatePath":"/etc/flyte/clusterresource/templates"}},"enabled":true,"templates":[{"key":"aa_namespace","value":"apiVersion: v1\nkind: Namespace\nmetadata:\n  name: {{ namespace }}\nspec:\n  finalizers:\n  - kubernetes\n"},{"key":"ab_project_resource_quota","value":"apiVersion: v1\nkind: ResourceQuota\nmetadata:\n  name: project-quota\n  namespace: {{ namespace }}\nspec:\n  hard:\n    limits.cpu: {{ projectQuotaCpu }}\n    limits.memory: {{ projectQuotaMemory }}\n"}]}` | Configuration for the Cluster resource manager component. This is an optional component, that enables automatic cluster configuration. This is useful to set default quotas, manage namespaces etc that map to a project/domain |
| flyte.cluster_resource_manager.config.cluster_resources | object | `{"customData":[{"production":[{"projectQuotaCpu":{"value":"5"}},{"projectQuotaMemory":{"value":"4000Mi"}}]},{"staging":[{"projectQuotaCpu":{"value":"2"}},{"projectQuotaMemory":{"value":"3000Mi"}}]},{"development":[{"projectQuotaCpu":{"value":"4"}},{"projectQuotaMemory":{"value":"3000Mi"}}]}],"refresh":"5m","refreshInterval":"5m","templatePath":"/etc/flyte/clusterresource/templates"}` | ClusterResource parameters Refer to the [structure](https://pkg.go.dev/github.com/lyft/flyteadmin@v0.3.37/pkg/runtime/interfaces#ClusterResourceConfig) to customize. |
| flyte.cluster_resource_manager.enabled | bool | `true` | Enables the Cluster resource manager component |
| flyte.cluster_resource_manager.templates | list | `[{"key":"aa_namespace","value":"apiVersion: v1\nkind: Namespace\nmetadata:\n  name: {{ namespace }}\nspec:\n  finalizers:\n  - kubernetes\n"},{"key":"ab_project_resource_quota","value":"apiVersion: v1\nkind: ResourceQuota\nmetadata:\n  name: project-quota\n  namespace: {{ namespace }}\nspec:\n  hard:\n    limits.cpu: {{ projectQuotaCpu }}\n    limits.memory: {{ projectQuotaMemory }}\n"}]` | Resource templates that should be applied |
| flyte.cluster_resource_manager.templates[0] | object | `{"key":"aa_namespace","value":"apiVersion: v1\nkind: Namespace\nmetadata:\n  name: {{ namespace }}\nspec:\n  finalizers:\n  - kubernetes\n"}` | Template for namespaces resources |
| flyte.common | object | `{"databaseSecret":{"name":"","secretManifest":{}},"flyteNamespaceTemplate":{"enabled":false},"ingress":{"albSSLRedirect":false,"annotations":{"nginx.ingress.kubernetes.io/app-root":"/console"},"enabled":true,"host":"","separateGrpcIngress":false,"separateGrpcIngressAnnotations":{"nginx.ingress.kubernetes.io/backend-protocol":"GRPC"},"tls":{"enabled":false},"webpackHMR":true}}` | ---------------------------------------------- COMMON SETTINGS |
| flyte.common.databaseSecret.name | string | `""` | Specify name of K8s Secret which contains Database password. Leave it empty if you don't need this Secret |
| flyte.common.databaseSecret.secretManifest | object | `{}` | Specify your Secret (with sensitive data) or pseudo-manifest (without sensitive data). See https://github.com/godaddy/kubernetes-external-secrets |
| flyte.common.flyteNamespaceTemplate.enabled | bool | `false` | - Enable or disable creating Flyte namespace in template. Enable when using helm as template-engine only. Disable when using `helm install ...`. |
| flyte.common.ingress.albSSLRedirect | bool | `false` | - albSSLRedirect adds a special route for ssl redirect. Only useful in combination with the AWS LoadBalancer Controller. |
| flyte.common.ingress.annotations | object | `{"nginx.ingress.kubernetes.io/app-root":"/console"}` | - Ingress annotations applied to both HTTP and GRPC ingresses. |
| flyte.common.ingress.enabled | bool | `true` | - Enable or disable creating Ingress for Flyte. Relevant to disable when using e.g. Istio as ingress controller. |
| flyte.common.ingress.host | string | `""` | - Ingress hostname |
| flyte.common.ingress.separateGrpcIngress | bool | `false` | - separateGrpcIngress puts GRPC routes into a separate ingress if true. Required for certain ingress controllers like nginx. |
| flyte.common.ingress.separateGrpcIngressAnnotations | object | `{"nginx.ingress.kubernetes.io/backend-protocol":"GRPC"}` | - Extra Ingress annotations applied only to the GRPC ingress. Only makes sense if `separateGrpcIngress` is enabled. |
| flyte.common.ingress.tls | object | `{"enabled":false}` | - TLS Settings |
| flyte.common.ingress.webpackHMR | bool | `true` | - Enable or disable HMR route to flyteconsole. This is useful only for frontend development. |
| flyte.configmap | object | `{"adminServer":{"auth":{"appAuth":{"thirdPartyConfig":{"flyteClient":{"clientId":"flytectl","redirectUri":"http://localhost:53593/callback","scopes":["offline","all"]}}},"authorizedUris":["https://localhost:30081","http://flyteadmin:80","http://flyteadmin.flyte.svc.cluster.local:80"],"userAuth":{"openId":{"baseUrl":"https://accounts.google.com","clientId":"657465813211-6eog7ek7li5k7i7fvgv2921075063hpe.apps.googleusercontent.com","scopes":["profile","openid"]}}},"flyteadmin":{"eventVersion":2,"metadataStoragePrefix":["metadata","admin"],"metricsScope":"flyte:","profilerPort":10254,"roleNameKey":"iam.amazonaws.com/role","testing":{"host":"http://flyteadmin"}},"server":{"grpcPort":8089,"httpPort":8088,"security":{"allowCors":true,"allowedHeaders":["Content-Type","flyte-authorization"],"allowedOrigins":["*"],"secure":false,"useAuth":false}}},"catalog":{"catalog-cache":{"endpoint":"datacatalog:89","insecure":true,"type":"datacatalog"}},"console":{"BASE_URL":"/console","CONFIG_DIR":"/etc/flyte/config","DISABLE_AUTH":"1"},"copilot":{"plugins":{"k8s":{"co-pilot":{"image":"cr.flyte.org/flyteorg/flytecopilot:v0.0.24","name":"flyte-copilot-","start-timeout":"30s"}}}},"core":{"propeller":{"downstream-eval-duration":"30s","enable-admin-launcher":true,"leader-election":{"enabled":true,"lease-duration":"15s","lock-config-map":{"name":"propeller-leader","namespace":"flyte"},"renew-deadline":"10s","retry-period":"2s"},"limit-namespace":"all","max-workflow-retries":30,"metadata-prefix":"metadata/propeller","metrics-prefix":"flyte","prof-port":10254,"queue":{"batch-size":-1,"batching-interval":"2s","queue":{"base-delay":"5s","capacity":1000,"max-delay":"120s","rate":100,"type":"maxof"},"sub-queue":{"capacity":100,"rate":10,"type":"bucket"},"type":"batch"},"rawoutput-prefix":"s3://my-s3-bucket/","workers":4,"workflow-reeval-duration":"30s"},"webhook":{"certDir":"/etc/webhook/certs","serviceName":"flyte-pod-webhook"}},"datacatalogServer":{"application":{"grpcPort":8089,"grpcServerReflection":true,"httpPort":8080},"datacatalog":{"metrics-scope":"datacatalog","profiler-port":10254,"storage-prefix":"metadata/datacatalog"}},"domain":{"domains":[{"id":"development","name":"development"},{"id":"staging","name":"staging"},{"id":"production","name":"production"}]},"enabled_plugins":{"tasks":{"task-plugins":{"default-for-task-types":{"container":"container","container_array":"k8s-array","sidecar":"sidecar"},"enabled-plugins":["container","sidecar","k8s-array"]}}},"k8s":{"plugins":{"k8s":{"default-cpus":"100m","default-env-vars":[{"FLYTE_AWS_ENDPOINT":"http://minio.flyte:9000"},{"FLYTE_AWS_ACCESS_KEY_ID":"minio"},{"FLYTE_AWS_SECRET_ACCESS_KEY":"miniostorage"}],"default-memory":"200Mi"}}},"logger":{"logger":{"level":4,"show-source":true}},"remoteData":{"remoteData":{"region":"us-east-1","scheme":"local","signedUrls":{"durationMinutes":3}}},"resource_manager":{"propeller":{"resourcemanager":{"redis":null,"type":"noop"}}},"task_logs":{"plugins":{"logs":{"cloudwatch-enabled":false,"kubernetes-enabled":true,"kubernetes-template-uri":"http://localhost:30082/#/log/{{ \"{{\" }} .namespace {{ \"}}\" }}/{{ \"{{\" }} .podName {{ \"}}\" }}/pod?namespace={{ \"{{\" }} .namespace {{ \"}}\" }}"}}},"task_resource_defaults":{"task_resources":{"defaults":{"cpu":"100m","memory":"200Mi","storage":"5Mi"},"limits":{"cpu":2,"gpu":1,"memory":"1Gi","storage":"20Mi"}}}}` | ----------------------------------------------------------------- CONFIGMAPS SETTINGS |
| flyte.configmap.adminServer | object | `{"auth":{"appAuth":{"thirdPartyConfig":{"flyteClient":{"clientId":"flytectl","redirectUri":"http://localhost:53593/callback","scopes":["offline","all"]}}},"authorizedUris":["https://localhost:30081","http://flyteadmin:80","http://flyteadmin.flyte.svc.cluster.local:80"],"userAuth":{"openId":{"baseUrl":"https://accounts.google.com","clientId":"657465813211-6eog7ek7li5k7i7fvgv2921075063hpe.apps.googleusercontent.com","scopes":["profile","openid"]}}},"flyteadmin":{"eventVersion":2,"metadataStoragePrefix":["metadata","admin"],"metricsScope":"flyte:","profilerPort":10254,"roleNameKey":"iam.amazonaws.com/role","testing":{"host":"http://flyteadmin"}},"server":{"grpcPort":8089,"httpPort":8088,"security":{"allowCors":true,"allowedHeaders":["Content-Type","flyte-authorization"],"allowedOrigins":["*"],"secure":false,"useAuth":false}}}` | FlyteAdmin server configuration |
| flyte.configmap.adminServer.auth | object | `{"appAuth":{"thirdPartyConfig":{"flyteClient":{"clientId":"flytectl","redirectUri":"http://localhost:53593/callback","scopes":["offline","all"]}}},"authorizedUris":["https://localhost:30081","http://flyteadmin:80","http://flyteadmin.flyte.svc.cluster.local:80"],"userAuth":{"openId":{"baseUrl":"https://accounts.google.com","clientId":"657465813211-6eog7ek7li5k7i7fvgv2921075063hpe.apps.googleusercontent.com","scopes":["profile","openid"]}}}` | Authentication configuration |
| flyte.configmap.adminServer.server.security.secure | bool | `false` | Controls whether to serve requests over SSL/TLS. |
| flyte.configmap.adminServer.server.security.useAuth | bool | `false` | Controls whether to enforce authentication. Follow the guide in https://docs.flyte.org/ on how to setup authentication. |
| flyte.configmap.catalog | object | `{"catalog-cache":{"endpoint":"datacatalog:89","insecure":true,"type":"datacatalog"}}` | Catalog Client configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/catalog#Config) Additional advanced Catalog configuration [here](https://pkg.go.dev/github.com/lyft/flyteplugins/go/tasks/pluginmachinery/catalog#Config) |
| flyte.configmap.console | object | `{"BASE_URL":"/console","CONFIG_DIR":"/etc/flyte/config","DISABLE_AUTH":"1"}` | Configuration for Flyte console UI |
| flyte.configmap.copilot | object | `{"plugins":{"k8s":{"co-pilot":{"image":"cr.flyte.org/flyteorg/flytecopilot:v0.0.24","name":"flyte-copilot-","start-timeout":"30s"}}}}` | Copilot configuration |
| flyte.configmap.copilot.plugins.k8s.co-pilot | object | `{"image":"cr.flyte.org/flyteorg/flytecopilot:v0.0.24","name":"flyte-copilot-","start-timeout":"30s"}` | Structure documented [here](https://pkg.go.dev/github.com/lyft/flyteplugins@v0.5.28/go/tasks/pluginmachinery/flytek8s/config#FlyteCoPilotConfig) |
| flyte.configmap.core | object | `{"propeller":{"downstream-eval-duration":"30s","enable-admin-launcher":true,"leader-election":{"enabled":true,"lease-duration":"15s","lock-config-map":{"name":"propeller-leader","namespace":"flyte"},"renew-deadline":"10s","retry-period":"2s"},"limit-namespace":"all","max-workflow-retries":30,"metadata-prefix":"metadata/propeller","metrics-prefix":"flyte","prof-port":10254,"queue":{"batch-size":-1,"batching-interval":"2s","queue":{"base-delay":"5s","capacity":1000,"max-delay":"120s","rate":100,"type":"maxof"},"sub-queue":{"capacity":100,"rate":10,"type":"bucket"},"type":"batch"},"rawoutput-prefix":"s3://my-s3-bucket/","workers":4,"workflow-reeval-duration":"30s"},"webhook":{"certDir":"/etc/webhook/certs","serviceName":"flyte-pod-webhook"}}` | Core propeller configuration |
| flyte.configmap.core.propeller | object | `{"downstream-eval-duration":"30s","enable-admin-launcher":true,"leader-election":{"enabled":true,"lease-duration":"15s","lock-config-map":{"name":"propeller-leader","namespace":"flyte"},"renew-deadline":"10s","retry-period":"2s"},"limit-namespace":"all","max-workflow-retries":30,"metadata-prefix":"metadata/propeller","metrics-prefix":"flyte","prof-port":10254,"queue":{"batch-size":-1,"batching-interval":"2s","queue":{"base-delay":"5s","capacity":1000,"max-delay":"120s","rate":100,"type":"maxof"},"sub-queue":{"capacity":100,"rate":10,"type":"bucket"},"type":"batch"},"rawoutput-prefix":"s3://my-s3-bucket/","workers":4,"workflow-reeval-duration":"30s"}` | follows the structure specified [here](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/config). |
| flyte.configmap.datacatalogServer | object | `{"application":{"grpcPort":8089,"grpcServerReflection":true,"httpPort":8080},"datacatalog":{"metrics-scope":"datacatalog","profiler-port":10254,"storage-prefix":"metadata/datacatalog"}}` | Datacatalog server config |
| flyte.configmap.domain | object | `{"domains":[{"id":"development","name":"development"},{"id":"staging","name":"staging"},{"id":"production","name":"production"}]}` | Domains configuration for Flyte projects. This enables the specified number of domains across all projects in Flyte. |
| flyte.configmap.enabled_plugins.tasks | object | `{"task-plugins":{"default-for-task-types":{"container":"container","container_array":"k8s-array","sidecar":"sidecar"},"enabled-plugins":["container","sidecar","k8s-array"]}}` | Tasks specific configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#GetConfig) |
| flyte.configmap.enabled_plugins.tasks.task-plugins | object | `{"default-for-task-types":{"container":"container","container_array":"k8s-array","sidecar":"sidecar"},"enabled-plugins":["container","sidecar","k8s-array"]}` | Plugins configuration, [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#TaskPluginConfig) |
| flyte.configmap.enabled_plugins.tasks.task-plugins.enabled-plugins | list | `["container","sidecar","k8s-array"]` | [Enabled Plugins](https://pkg.go.dev/github.com/lyft/flyteplugins/go/tasks/config#Config). Enable sagemaker*, athena if you install the backend plugins |
| flyte.configmap.k8s | object | `{"plugins":{"k8s":{"default-cpus":"100m","default-env-vars":[{"FLYTE_AWS_ENDPOINT":"http://minio.flyte:9000"},{"FLYTE_AWS_ACCESS_KEY_ID":"minio"},{"FLYTE_AWS_SECRET_ACCESS_KEY":"miniostorage"}],"default-memory":"200Mi"}}}` | Kubernetes specific Flyte configuration |
| flyte.configmap.k8s.plugins.k8s | object | `{"default-cpus":"100m","default-env-vars":[{"FLYTE_AWS_ENDPOINT":"http://minio.flyte:9000"},{"FLYTE_AWS_ACCESS_KEY_ID":"minio"},{"FLYTE_AWS_SECRET_ACCESS_KEY":"miniostorage"}],"default-memory":"200Mi"}` | Configuration section for all K8s specific plugins [Configuration structure](https://pkg.go.dev/github.com/lyft/flyteplugins/go/tasks/pluginmachinery/flytek8s/config) |
| flyte.configmap.logger | object | `{"logger":{"level":4,"show-source":true}}` | Logger configuration |
| flyte.configmap.resource_manager | object | `{"propeller":{"resourcemanager":{"redis":null,"type":"noop"}}}` | Resource manager configuration |
| flyte.configmap.resource_manager.propeller | object | `{"resourcemanager":{"redis":null,"type":"noop"}}` | resource manager configuration |
| flyte.configmap.task_logs | object | `{"plugins":{"logs":{"cloudwatch-enabled":false,"kubernetes-enabled":true,"kubernetes-template-uri":"http://localhost:30082/#/log/{{ \"{{\" }} .namespace {{ \"}}\" }}/{{ \"{{\" }} .podName {{ \"}}\" }}/pod?namespace={{ \"{{\" }} .namespace {{ \"}}\" }}"}}}` | Section that configures how the Task logs are displayed on the UI. This has to be changed based on your actual logging provider. Refer to [structure](https://pkg.go.dev/github.com/lyft/flyteplugins/go/tasks/logs#LogConfig) to understand how to configure various logging engines |
| flyte.configmap.task_logs.plugins.logs.cloudwatch-enabled | bool | `false` | One option is to enable cloudwatch logging for EKS, update the region and log group accordingly |
| flyte.configmap.task_resource_defaults | object | `{"task_resources":{"defaults":{"cpu":"100m","memory":"200Mi","storage":"5Mi"},"limits":{"cpu":2,"gpu":1,"memory":"1Gi","storage":"20Mi"}}}` | Task default resources configuration Refer to the full [structure](https://pkg.go.dev/github.com/lyft/flyteadmin@v0.3.37/pkg/runtime/interfaces#TaskResourceConfiguration). |
| flyte.configmap.task_resource_defaults.task_resources | object | `{"defaults":{"cpu":"100m","memory":"200Mi","storage":"5Mi"},"limits":{"cpu":2,"gpu":1,"memory":"1Gi","storage":"20Mi"}}` | Task default resources parameters |
| flyte.datacatalog.affinity | object | `{}` | affinity for Datacatalog deployment |
| flyte.datacatalog.configPath | string | `"/etc/datacatalog/config/*.yaml"` | Default regex string for searching configuration files |
| flyte.datacatalog.image.pullPolicy | string | `"IfNotPresent"` | Docker image pull policy |
| flyte.datacatalog.image.repository | string | `"cr.flyte.org/flyteorg/datacatalog"` | Docker image for Datacatalog deployment |
| flyte.datacatalog.image.tag | string | `"v0.3.19"` | Docker image tag |
| flyte.datacatalog.nodeSelector | object | `{}` | nodeSelector for Datacatalog deployment |
| flyte.datacatalog.podAnnotations | object | `{}` | Annotations for Datacatalog pods |
| flyte.datacatalog.replicaCount | int | `1` | Replicas count for Datacatalog deployment |
| flyte.datacatalog.resources | object | `{"limits":{"cpu":"500m","ephemeral-storage":"100Mi","memory":"500Mi"},"requests":{"cpu":"10m","ephemeral-storage":"50Mi","memory":"50Mi"}}` | Default resources requests and limits for Datacatalog deployment |
| flyte.datacatalog.service | object | `{"annotations":{"projectcontour.io/upstream-protocol.h2c":"grpc"},"type":"NodePort"}` | Service settings for Datacatalog |
| flyte.datacatalog.serviceAccount | object | `{"annotations":{},"create":true,"imagePullSecrets":{}}` | Configuration for service accounts for Datacatalog |
| flyte.datacatalog.serviceAccount.annotations | object | `{}` | Annotations for ServiceAccount attached to Datacatalog pods |
| flyte.datacatalog.serviceAccount.create | bool | `true` | Should a service account be created for Datacatalog |
| flyte.datacatalog.serviceAccount.imagePullSecrets | object | `{}` | ImapgePullSecrets to automatically assign to the service account |
| flyte.datacatalog.tolerations | list | `[]` | tolerations for Datacatalog deployment |
| flyte.flyteadmin.affinity | object | `{}` | affinity for Flyteadmin deployment |
| flyte.flyteadmin.configPath | string | `"/etc/flyte/config/*.yaml"` | Default regex string for searching configuration files |
| flyte.flyteadmin.deployRedoc | bool | `true` | Deploys a Redoc container in Flyteadmin's pod |
| flyte.flyteadmin.image.pullPolicy | string | `"IfNotPresent"` | Docker image pull policy |
| flyte.flyteadmin.image.repository | string | `"cr.flyte.org/flyteorg/flyteadmin"` | Docker image for Flyteadmin deployment |
| flyte.flyteadmin.image.tag | string | `"v0.6.89"` | Docker image tag |
| flyte.flyteadmin.initialProjects | list | `["flytesnacks","flytetester","flyteexamples"]` | Initial projects to create |
| flyte.flyteadmin.nodeSelector | object | `{}` | nodeSelector for Flyteadmin deployment |
| flyte.flyteadmin.podAnnotations | object | `{}` | Annotations for Flyteadmin pods |
| flyte.flyteadmin.replicaCount | int | `1` | Replicas count for Flyteadmin deployment |
| flyte.flyteadmin.resources | object | `{"limits":{"cpu":"250m","ephemeral-storage":"100Mi","memory":"500Mi"},"requests":{"cpu":"10m","ephemeral-storage":"50Mi","memory":"50Mi"}}` | Default resources requests and limits for Flyteadmin deployment |
| flyte.flyteadmin.service | object | `{"annotations":{"projectcontour.io/upstream-protocol.h2c":"grpc"},"loadBalancerSourceRanges":[],"type":"ClusterIP"}` | Service settings for Flyteadmin |
| flyte.flyteadmin.serviceAccount | object | `{"annotations":{},"create":true,"imagePullSecrets":{}}` | Configuration for service accounts for FlyteAdmin |
| flyte.flyteadmin.serviceAccount.annotations | object | `{}` | Annotations for ServiceAccount attached to Flyteadmin pods |
| flyte.flyteadmin.serviceAccount.create | bool | `true` | Should a service account be created for flyteadmin |
| flyte.flyteadmin.serviceAccount.imagePullSecrets | object | `{}` | ImapgePullSecrets to automatically assign to the service account |
| flyte.flyteadmin.tolerations | list | `[]` | tolerations for Flyteadmin deployment |
| flyte.flyteconsole.affinity | object | `{}` | affinity for Flyteconsole deployment |
| flyte.flyteconsole.image.pullPolicy | string | `"IfNotPresent"` | Docker image pull policy |
| flyte.flyteconsole.image.repository | string | `"cr.flyte.org/flyteorg/flyteconsole"` | Docker image for Flyteconsole deployment |
| flyte.flyteconsole.image.tag | string | `"v0.40.0"` | Docker image tag |
| flyte.flyteconsole.nodeSelector | object | `{}` | nodeSelector for Flyteconsole deployment |
| flyte.flyteconsole.podAnnotations | object | `{}` | Annotations for Flyteconsole pods |
| flyte.flyteconsole.replicaCount | int | `1` | Replicas count for Flyteconsole deployment |
| flyte.flyteconsole.resources | object | `{"limits":{"cpu":"500m","memory":"250Mi"},"requests":{"cpu":"10m","memory":"50Mi"}}` | Default resources requests and limits for Flyteconsole deployment |
| flyte.flyteconsole.service | object | `{"annotations":{},"type":"ClusterIP"}` | Service settings for Flyteconsole |
| flyte.flyteconsole.tolerations | list | `[]` | tolerations for Flyteconsole deployment |
| flyte.flytepropeller.affinity | object | `{}` | affinity for Flytepropeller deployment |
| flyte.flytepropeller.configPath | string | `"/etc/flyte/config/*.yaml"` | Default regex string for searching configuration files |
| flyte.flytepropeller.image.pullPolicy | string | `"IfNotPresent"` | Docker image pull policy |
| flyte.flytepropeller.image.repository | string | `"cr.flyte.org/flyteorg/flytepropeller"` | Docker image for Flytepropeller deployment |
| flyte.flytepropeller.image.tag | string | `"v0.16.14"` | Docker image tag |
| flyte.flytepropeller.nodeSelector | object | `{}` | nodeSelector for Flytepropeller deployment |
| flyte.flytepropeller.podAnnotations | object | `{}` | Annotations for Flytepropeller pods |
| flyte.flytepropeller.replicaCount | int | `1` | Replicas count for Flytepropeller deployment |
| flyte.flytepropeller.resources | object | `{"limits":{"cpu":"200m","ephemeral-storage":"100Mi","memory":"200Mi"},"requests":{"cpu":"10m","ephemeral-storage":"50Mi","memory":"50Mi"}}` | Default resources requests and limits for Flytepropeller deployment |
| flyte.flytepropeller.serviceAccount | object | `{"annotations":{},"create":true,"imagePullSecrets":{}}` | Configuration for service accounts for FlytePropeller |
| flyte.flytepropeller.serviceAccount.annotations | object | `{}` | Annotations for ServiceAccount attached to FlytePropeller pods |
| flyte.flytepropeller.serviceAccount.create | bool | `true` | Should a service account be created for FlytePropeller |
| flyte.flytepropeller.serviceAccount.imagePullSecrets | object | `{}` | ImapgePullSecrets to automatically assign to the service account |
| flyte.flytepropeller.tolerations | list | `[]` | tolerations for Flytepropeller deployment |
| flyte.flytescheduler.affinity | object | `{}` | affinity for Flytescheduler deployment |
| flyte.flytescheduler.configPath | string | `"/etc/flyte/config/*.yaml"` | Default regex string for searching configuration files |
| flyte.flytescheduler.image.pullPolicy | string | `"IfNotPresent"` | Docker image pull policy |
| flyte.flytescheduler.image.repository | string | `"cr.flyte.org/flyteorg/flytescheduler"` | Docker image for Flytescheduler deployment |
| flyte.flytescheduler.image.tag | string | `"v0.6.89"` | Docker image tag |
| flyte.flytescheduler.nodeSelector | object | `{}` | nodeSelector for Flytescheduler deployment |
| flyte.flytescheduler.podAnnotations | object | `{}` | Annotations for Flytescheduler pods |
| flyte.flytescheduler.resources | object | `{"limits":{"cpu":"250m","ephemeral-storage":"100Mi","memory":"500Mi"},"requests":{"cpu":"10m","ephemeral-storage":"50Mi","memory":"50Mi"}}` | Default resources requests and limits for Flytescheduler deployment |
| flyte.flytescheduler.serviceAccount | object | `{"annotations":{},"create":true,"imagePullSecrets":{}}` | Configuration for service accounts for Flytescheduler |
| flyte.flytescheduler.serviceAccount.annotations | object | `{}` | Annotations for ServiceAccount attached to Flytescheduler pods |
| flyte.flytescheduler.serviceAccount.create | bool | `true` | Should a service account be created for Flytescheduler |
| flyte.flytescheduler.serviceAccount.imagePullSecrets | object | `{}` | ImapgePullSecrets to automatically assign to the service account |
| flyte.flytescheduler.tolerations | list | `[]` | tolerations for Flytescheduler deployment |
| flyte.storage | object | `{"bucketName":"my-s3-bucket","custom":{},"gcs":null,"s3":{"region":"us-east-1"},"type":"sandbox"}` | ---------------------------------------------------- STORAGE SETTINGS |
| flyte.storage.bucketName | string | `"my-s3-bucket"` | bucketName defines the storage bucket flyte will use. Required for all types except for sandbox. |
| flyte.storage.custom | object | `{}` | Settings for storage type custom. See https://github:com/graymeta/stow for supported storage providers/settings. |
| flyte.storage.gcs | string | `nil` | settings for storage type gcs |
| flyte.storage.s3 | object | `{"region":"us-east-1"}` | settings for storage type s3 |
| flyte.storage.type | string | `"sandbox"` | Sets the storage type. Supported values are sandbox, s3, gcs and custom. |
| flyte.webhook.enabled | bool | `true` | enable or disable secrets webhook |
| flyte.webhook.service | object | `{"annotations":{"projectcontour.io/upstream-protocol.h2c":"grpc"},"type":"ClusterIP"}` | Service settings for the webhook |
| flyte.webhook.serviceAccount | object | `{"annotations":{},"create":true,"imagePullSecrets":{}}` | Configuration for service accounts for the webhook |
| flyte.webhook.serviceAccount.annotations | object | `{}` | Annotations for ServiceAccount attached to the webhook |
| flyte.webhook.serviceAccount.create | bool | `true` | Should a service account be created for the webhook |
| flyte.webhook.serviceAccount.imagePullSecrets | object | `{}` | ImapgePullSecrets to automatically assign to the service account |
| flyte.workflow_notifications | object | `{"config":{},"enabled":false}` | **Optional Component** Workflow notifications module is an optional dependency. Flyte uses cloud native pub-sub systems to notify users of various events in their workflows |
| flyte.workflow_scheduler | object | `{"enabled":true,"type":"native"}` | **Optional Component** Flyte uses a cloud hosted Cron scheduler to run workflows on a schedule. The following module is optional. Without, this module, you will not have scheduled launchplans / workflows. Docs: https://docs.flyte.org/en/latest/howto/enable_and_use_schedules.html#setting-up-scheduled-workflows |
| kubernetes-dashboard.enabled | bool | `true` |  |
| kubernetes-dashboard.extraArgs[0] | string | `"--enable-skip-login"` |  |
| kubernetes-dashboard.extraArgs[1] | string | `"--enable-insecure-login"` |  |
| kubernetes-dashboard.extraArgs[2] | string | `"--disable-settings-authorizer"` |  |
| kubernetes-dashboard.protocolHttp | bool | `true` |  |
| kubernetes-dashboard.rbac.clusterReadOnlyRole | bool | `true` |  |
| kubernetes-dashboard.service.externalPort | int | `30082` |  |
| kubernetes-dashboard.service.nodePort | int | `30082` |  |
| kubernetes-dashboard.service.type | string | `"NodePort"` |  |
| minio.affinity | object | `{}` | affinity for Minio deployment |
| minio.enabled | bool | `true` | - enable or disable Minio deployment installation |
| minio.image.pullPolicy | string | `"IfNotPresent"` | Docker image pull policy |
| minio.image.repository | string | `"ecr.flyte.org/bitnami/minio"` | Docker image for Minio deployment |
| minio.image.tag | string | `"2021.10.13-debian-10-r0"` | Docker image tag |
| minio.nodeSelector | object | `{}` | nodeSelector for Minio deployment |
| minio.podAnnotations | object | `{}` | Annotations for Minio pods |
| minio.replicaCount | int | `1` | Replicas count for Minio deployment |
| minio.resources | object | `{"limits":{"cpu":"200m","memory":"512Mi"},"requests":{"cpu":"10m","memory":"128Mi"}}` | Default resources requests and limits for Minio deployment |
| minio.resources.limits | object | `{"cpu":"200m","memory":"512Mi"}` | Limits are the maximum set of resources needed for this pod |
| minio.resources.requests | object | `{"cpu":"10m","memory":"128Mi"}` | Requests are the minimum set of resources needed for this pod |
| minio.service | object | `{"annotations":{},"type":"NodePort"}` | Service settings for Minio |
| minio.tolerations | list | `[]` | tolerations for Minio deployment |
| postgres.affinity | object | `{}` | affinity for Postgres deployment |
| postgres.enabled | bool | `true` | - enable or disable Postgres deployment installation |
| postgres.image.pullPolicy | string | `"IfNotPresent"` | Docker image pull policy |
| postgres.image.repository | string | `"ecr.flyte.org/ubuntu/postgres"` | Docker image for Postgres deployment |
| postgres.image.tag | string | `"13-21.04_beta"` | Docker image tag |
| postgres.nodeSelector | object | `{}` | nodeSelector for Postgres deployment |
| postgres.podAnnotations | object | `{}` | Annotations for Postgres pods |
| postgres.replicaCount | int | `1` | Replicas count for Postgres deployment |
| postgres.resources | object | `{"limits":{"cpu":"1000m","memory":"512Mi"},"requests":{"cpu":"10m","memory":"128Mi"}}` | Default resources requests and limits for Postgres deployment |
| postgres.service | object | `{"annotations":{},"type":"ClusterIP"}` | Service settings for Postgres |
| postgres.tolerations | list | `[]` | tolerations for Postgres deployment |
| redis | object | `{"enabled":false}` | --------------------------------------------- REDIS SETTINGS |
| redis.enabled | bool | `false` | - enable or disable Redis Statefulset installation |
| sparkoperator | object | `{"enabled":false}` | Optional: Spark Plugin using the Spark Operator |
| sparkoperator.enabled | bool | `false` | - enable or disable Sparkoperator deployment installation |
