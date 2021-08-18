# flyte-sandbox

![Version: v0.1.10](https://img.shields.io/badge/Version-v0.1.10-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)

A Helm chart for Flyte

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../flyte | flyte | v0.1.10 |
| https://charts.bitnami.com/bitnami | contour | 4.1.2 |
| https://googlecloudplatform.github.io/spark-on-k8s-operator | sparkoperator(spark-operator) | 1.0.6 |
| https://kubernetes.github.io/dashboard/ | kubernetes-dashboard | 4.0.2 |

### SANDBOX INSTALLATION:
- [Install helm 3](https://helm.sh/docs/intro/install/)
- Fetch chart dependencies ``
- Install Flyte sandbox:

```bash
helm repo add flyte https://flyteorg.github.io/flyte
helm dep up
helm install -n flyte -f values.yaml --create-namespace flyte flyte/flyte-sandbox
```

Customize your installation by changing settings in `values-sandbox.yaml`.
You can use the helm diff plugin to review any value changes you've made to your values:

```bash
helm plugin install https://github.com/databus23/helm-diff
helm diff upgrade -f values.yaml flyte flyte/flyte-sandbox
```

Then apply your changes:
```bash
helm upgrade -f values.yaml flyte flyte/flyte-sandbox
```

#### Alternative: Generate raw kubernetes yaml with helm template
- `helm template --name-template=flyte-sandbox flyte/flyte-sandbox -n flyte -f values.yaml > flyte_generated_sandbox.yaml`
- Deploy the manifest `kubectl apply -f flyte_generated_sandbox.yaml`

- When all pods are running - run end2end tests: `kubectl apply -f ../end2end/tests/endtoend.yaml`
- Get flyte host `minikube service contour -n heptio-contour --url`. And then visit `http://<HOST>/console`

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| cluster_resource_manager | object | `{"enabled":true}` | Configuration for the Cluster resource manager component. This is an optional component, that enables automatic cluster configuration. This is useful to set default quotas, manage namespaces etc that map to a project/domain |
| cluster_resource_manager.enabled | bool | `true` | Enables the Cluster resource manager component |
| common.databaseSecret | object | `{}` |  |
| common.flyteNamespaceTemplate | object | `{}` |  |
| common.ingress.albSSLRedirect | bool | `false` |  |
| common.ingress.annotations | object | `{}` |  |
| common.ingress.enabled | bool | `true` |  |
| common.ingress.separateGrpcIngress | bool | `false` |  |
| common.ingress.separateGrpcIngressAnnotations."nginx.ingress.kubernetes.io/backend-protocol" | string | `"GRPC"` |  |
| common.ingress.tls.enabled | bool | `false` |  |
| common.ingress.webpackHMR | bool | `true` |  |
| configmap.admin.admin.clientId | string | `"flytepropeller"` |  |
| configmap.admin.admin.clientSecretLocation | string | `"/etc/secrets/client_secret"` |  |
| configmap.admin.admin.endpoint | string | `"flyteadmin:81"` |  |
| configmap.admin.admin.insecure | bool | `true` |  |
| configmap.admin.event.capacity | int | `1000` |  |
| configmap.admin.event.rate | int | `500` |  |
| configmap.admin.event.type | string | `"admin"` |  |
| configmap.adminServer | object | `{"auth":{"appAuth":{"thirdPartyConfig":{"flyteClient":{"clientId":"flytectl","redirectUri":"http://localhost:53593/callback","scopes":["offline","all"]}}},"authorizedUris":["https://localhost:30081","http://flyteadmin:80","http://flyteadmin.flyte.svc.cluster.local:80"],"userAuth":{"openId":{"baseUrl":"https://accounts.google.com","clientId":"657465813211-6eog7ek7li5k7i7fvgv2921075063hpe.apps.googleusercontent.com","scopes":["profile","openid"]}}},"flyteadmin":{"eventVersion":1,"metadataStoragePrefix":["metadata","admin"],"metricsScope":"flyte:","profilerPort":10254,"roleNameKey":"iam.amazonaws.com/role","testing":{"host":"http://flyteadmin"}},"server":{"grpcPort":8089,"httpPort":8088,"security":{"allowCors":true,"allowedHeaders":["Content-Type","flyte-authorization"],"allowedOrigins":["*"],"secure":false,"useAuth":false}}}` | FlyteAdmin server configuration |
| configmap.adminServer.auth | object | `{"appAuth":{"thirdPartyConfig":{"flyteClient":{"clientId":"flytectl","redirectUri":"http://localhost:53593/callback","scopes":["offline","all"]}}},"authorizedUris":["https://localhost:30081","http://flyteadmin:80","http://flyteadmin.flyte.svc.cluster.local:80"],"userAuth":{"openId":{"baseUrl":"https://accounts.google.com","clientId":"657465813211-6eog7ek7li5k7i7fvgv2921075063hpe.apps.googleusercontent.com","scopes":["profile","openid"]}}}` | Authentication configuration |
| configmap.adminServer.server.security.secure | bool | `false` | Controls whether to serve requests over SSL/TLS. |
| configmap.adminServer.server.security.useAuth | bool | `false` | Controls whether to enforce authentication. Follow the guide in https://docs.flyte.org/ on how to setup authentication. |
| configmap.catalog | object | `{"catalog-cache":{"endpoint":"datacatalog:89","insecure":true,"type":"datacatalog"}}` | Catalog Client configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/catalog#Config) Additional advanced Catalog configuration [here](https://pkg.go.dev/github.com/lyft/flyteplugins/go/tasks/pluginmachinery/catalog#Config) |
| configmap.console | object | `{"BASE_URL":"/console","CONFIG_DIR":"/etc/flyte/config","DISABLE_AUTH":"1"}` | Configuration for Flyte console UI |
| configmap.copilot | object | `{"plugins":{"k8s":{"co-pilot":{"image":"cr.flyte.org/lyft/flyteplugins/flytecopilot:dc4bdbd61cac88a39a5ff43e40f026bdbc2c78a2","name":"flyte-copilot-","start-timeout":"30s"}}}}` | Copilot configuration |
| configmap.copilot.plugins.k8s.co-pilot | object | `{"image":"cr.flyte.org/lyft/flyteplugins/flytecopilot:dc4bdbd61cac88a39a5ff43e40f026bdbc2c78a2","name":"flyte-copilot-","start-timeout":"30s"}` | Structure documented [here](https://pkg.go.dev/github.com/lyft/flyteplugins@v0.5.28/go/tasks/pluginmachinery/flytek8s/config#FlyteCoPilotConfig) |
| configmap.core | object | `{"propeller":{"downstream-eval-duration":"30s","enable-admin-launcher":true,"leader-election":{"enabled":true,"lease-duration":"15s","lock-config-map":{"name":"propeller-leader","namespace":"flyte"},"renew-deadline":"10s","retry-period":"2s"},"limit-namespace":"all","max-workflow-retries":30,"metadata-prefix":"metadata/propeller","metrics-prefix":"flyte","prof-port":10254,"queue":{"batch-size":-1,"batching-interval":"2s","queue":{"base-delay":"5s","capacity":1000,"max-delay":"120s","rate":100,"type":"maxof"},"sub-queue":{"capacity":100,"rate":10,"type":"bucket"},"type":"batch"},"rawoutput-prefix":"s3://my-s3-bucket/","workers":20,"workflow-reeval-duration":"30s"},"webhook":{"certDir":"/etc/webhook/certs","serviceName":"flyte-pod-webhook"}}` | Core propeller configuration |
| configmap.core.propeller | object | `{"downstream-eval-duration":"30s","enable-admin-launcher":true,"leader-election":{"enabled":true,"lease-duration":"15s","lock-config-map":{"name":"propeller-leader","namespace":"flyte"},"renew-deadline":"10s","retry-period":"2s"},"limit-namespace":"all","max-workflow-retries":30,"metadata-prefix":"metadata/propeller","metrics-prefix":"flyte","prof-port":10254,"queue":{"batch-size":-1,"batching-interval":"2s","queue":{"base-delay":"5s","capacity":1000,"max-delay":"120s","rate":100,"type":"maxof"},"sub-queue":{"capacity":100,"rate":10,"type":"bucket"},"type":"batch"},"rawoutput-prefix":"s3://my-s3-bucket/","workers":20,"workflow-reeval-duration":"30s"}` | follows the structure specified [here](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/config). |
| configmap.datacatalogServer | object | `{"application":{"grpcPort":8089,"grpcServerReflection":true,"httpPort":8080},"datacatalog":{"metrics-scope":"datacatalog","profiler-port":10254,"storage-prefix":"metadata/datacatalog"}}` | Datacatalog server config |
| configmap.domain | object | `{"domains":[{"id":"development","name":"development"},{"id":"staging","name":"staging"},{"id":"production","name":"production"}]}` | Domains configuration for Flyte projects. This enables the specified number of domains across all projects in Flyte. |
| configmap.enabled_plugins.tasks | object | `{"task-plugins":{"default-for-task-types":{"container":"container","container_array":"k8s-array","sidecar":"sidecar"},"enabled-plugins":["container","sidecar","k8s-array"]}}` | Tasks specific configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#GetConfig) |
| configmap.enabled_plugins.tasks.task-plugins | object | `{"default-for-task-types":{"container":"container","container_array":"k8s-array","sidecar":"sidecar"},"enabled-plugins":["container","sidecar","k8s-array"]}` | Plugins configuration, [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#TaskPluginConfig) |
| configmap.enabled_plugins.tasks.task-plugins.enabled-plugins | list | `["container","sidecar","k8s-array"]` | [Enabled Plugins](https://pkg.go.dev/github.com/lyft/flyteplugins/go/tasks/config#Config). Enable sagemaker*, athena if you install the backend plugins |
| configmap.k8s | object | `{"plugins":{"k8s":{"default-cpus":"100m","default-env-vars":[{"FLYTE_AWS_ENDPOINT":"http://minio.flyte:9000"},{"FLYTE_AWS_ACCESS_KEY_ID":"minio"},{"FLYTE_AWS_SECRET_ACCESS_KEY":"miniostorage"}],"default-memory":"200Mi"}}}` | Kubernetes specific Flyte configuration |
| configmap.k8s.plugins.k8s | object | `{"default-cpus":"100m","default-env-vars":[{"FLYTE_AWS_ENDPOINT":"http://minio.flyte:9000"},{"FLYTE_AWS_ACCESS_KEY_ID":"minio"},{"FLYTE_AWS_SECRET_ACCESS_KEY":"miniostorage"}],"default-memory":"200Mi"}` | Configuration section for all K8s specific plugins [Configuration structure](https://pkg.go.dev/github.com/lyft/flyteplugins/go/tasks/pluginmachinery/flytek8s/config) |
| configmap.logger | object | `{"logger":{"level":4,"show-source":true}}` | Logger configuration |
| configmap.remoteData.remoteData.region | string | `"us-east-1"` |  |
| configmap.remoteData.remoteData.scheme | string | `"local"` |  |
| configmap.remoteData.remoteData.signedUrls.durationMinutes | int | `3` |  |
| configmap.resource_manager | object | `{"propeller":{"resourcemanager":{"redis":null,"type":"noop"}}}` | Resource manager configuration |
| configmap.resource_manager.propeller | object | `{"resourcemanager":{"redis":null,"type":"noop"}}` | resource manager configuration |
| configmap.task_logs | object | `{"plugins":{"logs":{"kubernetes-enabled":true,"kubernetes-template-uri":"http://localhost:30082/#/log/{{ .namespace }}/{{ .podName }}/pod?namespace={{ .namespace }}"}}}` | Section that configures how the Task logs are displayed on the UI. This has to be changed based on your actual logging provider. Refer to [structure](https://pkg.go.dev/github.com/lyft/flyteplugins/go/tasks/logs#LogConfig) to understand how to configure various logging engines |
| configmap.task_resource_defaults | object | `{"task_resources":{"defaults":{"cpu":"100m","memory":"200Mi","storage":"5Mi"},"limits":{"cpu":2,"gpu":1,"memory":"8Gi","storage":"20Mi"}}}` | Task default resources configuration Refer to the full [structure](https://pkg.go.dev/github.com/lyft/flyteadmin@v0.3.37/pkg/runtime/interfaces#TaskResourceConfiguration). |
| configmap.task_resource_defaults.task_resources | object | `{"defaults":{"cpu":"100m","memory":"200Mi","storage":"5Mi"},"limits":{"cpu":2,"gpu":1,"memory":"8Gi","storage":"20Mi"}}` | Task default resources parameters |
| contour.contour.resources | object | `{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"10m","memory":"50Mi"}}` | Default resources requests and limits for Contour |
| contour.enabled | bool | `true` |  |
| contour.envoy.resources | object | `{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"10m","memory":"50Mi"}}` | Default resources requests and limits for Envoy |
| contour.envoy.service.nodePorts.http | int | `30081` |  |
| contour.envoy.service.ports.http | int | `80` |  |
| contour.envoy.service.type | string | `"NodePort"` |  |
| contour.nodeSelector | object | `{}` | nodeSelector for Contour deployment |
| contour.podAnnotations | object | `{}` | Annotations for Contour pods |
| contour.replicaCount | int | `1` | Replicas count for Contour deployment |
| contour.serviceAccountAnnotations | object | `{}` | Annotations for ServiceAccount attached to Contour pods |
| contour.tolerations | list | `[]` | tolerations for Contour deployment |
| datacatalog.affinity | object | `{}` | affinity for Datacatalog deployment |
| datacatalog.configPath | string | `"/etc/datacatalog/config/*.yaml"` | Default regex string for searching configuration files |
| datacatalog.image.pullPolicy | string | `"IfNotPresent"` |  |
| datacatalog.image.repository | string | `"cr.flyte.org/flyteorg/datacatalog"` | Docker image for Datacatalog deployment |
| datacatalog.image.tag | string | `"v0.3.5"` |  |
| datacatalog.nodeSelector | object | `{}` | nodeSelector for Datacatalog deployment |
| datacatalog.podAnnotations | object | `{}` | Annotations for Datacatalog pods |
| datacatalog.replicaCount | int | `1` | Replicas count for Datacatalog deployment |
| datacatalog.resources | object | `{"limits":{"cpu":"500m","ephemeral-storage":"100Mi","memory":"500Mi"},"requests":{"cpu":"10m","ephemeral-storage":"50Mi","memory":"50Mi"}}` | Default resources requests and limits for Datacatalog deployment |
| datacatalog.service | object | `{"annotations":{"projectcontour.io/upstream-protocol.h2c":"grpc"},"type":"NodePort"}` | Service settings for Datacatalog |
| datacatalog.serviceAccount | object | `{"annotations":{},"create":true,"imagePullSecrets":{}}` | Configuration for service accounts for Datacatalog |
| datacatalog.serviceAccount.annotations | object | `{}` | Annotations for ServiceAccount attached to Datacatalog pods |
| datacatalog.serviceAccount.create | bool | `true` | Should a service account be created for Datacatalog |
| datacatalog.serviceAccount.imagePullSecrets | object | `{}` | ImapgePullSecrets to automatically assign to the service account |
| datacatalog.tolerations | list | `[]` | tolerations for Datacatalog deployment |
| db.database.dbname | string | `"flyte_development"` |  |
| db.database.host | string | `"postgres"` |  |
| db.database.port | int | `5432` |  |
| db.database.username | string | `"postgres"` |  |
| flyteadmin.affinity | object | `{}` | affinity for Flyteadmin deployment |
| flyteadmin.configPath | string | `"/etc/flyte/config/*.yaml"` | Default regex string for searching configuration files |
| flyteadmin.image.pullPolicy | string | `"IfNotPresent"` |  |
| flyteadmin.image.repository | string | `"cr.flyte.org/flyteorg/flyteadmin"` | Docker image for Flyteadmin deployment |
| flyteadmin.image.tag | string | `"v0.6.10"` |  |
| flyteadmin.initialProjects | list | `["flytesnacks","flytetester","flyteexamples"]` | Initial projects to create |
| flyteadmin.nodeSelector | object | `{}` | nodeSelector for Flyteadmin deployment |
| flyteadmin.podAnnotations | object | `{}` | Annotations for Flyteadmin pods |
| flyteadmin.replicaCount | int | `1` | Replicas count for Flyteadmin deployment |
| flyteadmin.resources | object | `{"limits":{"cpu":"250m","ephemeral-storage":"100Mi","memory":"500Mi"},"requests":{"cpu":"10m","ephemeral-storage":"50Mi","memory":"50Mi"}}` | Default resources requests and limits for Flyteadmin deployment |
| flyteadmin.secrets | object | `{}` |  |
| flyteadmin.service | object | `{"annotations":{"projectcontour.io/upstream-protocol.h2c":"grpc"},"loadBalancerSourceRanges":[],"type":"ClusterIP"}` | Service settings for Flyteadmin |
| flyteadmin.serviceAccount | object | `{"annotations":{},"create":true,"imagePullSecrets":{}}` | Configuration for service accounts for FlyteAdmin |
| flyteadmin.serviceAccount.annotations | object | `{}` | Annotations for ServiceAccount attached to Flyteadmin pods |
| flyteadmin.serviceAccount.create | bool | `true` | Should a service account be created for flyteadmin |
| flyteadmin.serviceAccount.imagePullSecrets | object | `{}` | ImapgePullSecrets to automatically assign to the service account |
| flyteadmin.tolerations | list | `[]` | tolerations for Flyteadmin deployment |
| flyteconsole.affinity | object | `{}` | affinity for Flyteconsole deployment |
| flyteconsole.image.pullPolicy | string | `"IfNotPresent"` |  |
| flyteconsole.image.repository | string | `"cr.flyte.org/flyteorg/flyteconsole"` | Docker image for Flyteconsole deployment |
| flyteconsole.image.tag | string | `"v0.20.1"` |  |
| flyteconsole.nodeSelector | object | `{}` | nodeSelector for Flyteconsole deployment |
| flyteconsole.podAnnotations | object | `{}` | Annotations for Flyteconsole pods |
| flyteconsole.replicaCount | int | `1` | Replicas count for Flyteconsole deployment |
| flyteconsole.resources | object | `{"limits":{"cpu":"500m","memory":"250Mi"},"requests":{"cpu":"10m","memory":"50Mi"}}` | Default resources requests and limits for Flyteconsole deployment |
| flyteconsole.service | object | `{"annotations":{},"type":"ClusterIP"}` | Service settings for Flyteconsole |
| flyteconsole.tolerations | list | `[]` | tolerations for Flyteconsole deployment |
| flytepropeller.affinity | object | `{}` | affinity for Flytepropeller deployment |
| flytepropeller.cacheSizeMbs | int | `0` |  |
| flytepropeller.configPath | string | `"/etc/flyte/config/*.yaml"` | Default regex string for searching configuration files |
| flytepropeller.image.pullPolicy | string | `"IfNotPresent"` |  |
| flytepropeller.image.repository | string | `"cr.flyte.org/flyteorg/flytepropeller"` | Docker image for Flytepropeller deployment |
| flytepropeller.image.tag | string | `"v0.12.9"` |  |
| flytepropeller.nodeSelector | object | `{}` | nodeSelector for Flytepropeller deployment |
| flytepropeller.podAnnotations | object | `{}` | Annotations for Flytepropeller pods |
| flytepropeller.replicaCount | int | `1` | Replicas count for Flytepropeller deployment |
| flytepropeller.resources | object | `{"limits":{"cpu":"200m","ephemeral-storage":"100Mi","memory":"200Mi"},"requests":{"cpu":"10m","ephemeral-storage":"50Mi","memory":"50Mi"}}` | Default resources requests and limits for Flytepropeller deployment |
| flytepropeller.serviceAccount | object | `{"annotations":{},"create":true,"imagePullSecrets":{}}` | Configuration for service accounts for FlytePropeller |
| flytepropeller.serviceAccount.annotations | object | `{}` | Annotations for ServiceAccount attached to FlytePropeller pods |
| flytepropeller.serviceAccount.create | bool | `true` | Should a service account be created for FlytePropeller |
| flytepropeller.serviceAccount.imagePullSecrets | object | `{}` | ImapgePullSecrets to automatically assign to the service account |
| flytepropeller.tolerations | list | `[]` | tolerations for Flytepropeller deployment |
| kubernetes-dashboard.enabled | bool | `true` |  |
| kubernetes-dashboard.extraArgs[0] | string | `"--enable-skip-login"` |  |
| kubernetes-dashboard.extraArgs[1] | string | `"--enable-insecure-login"` |  |
| kubernetes-dashboard.extraArgs[2] | string | `"--disable-settings-authorizer"` |  |
| kubernetes-dashboard.protocolHttp | bool | `true` |  |
| kubernetes-dashboard.service.externalPort | int | `30082` |  |
| kubernetes-dashboard.service.type | string | `"NodePort"` |  |
| minio.affinity | object | `{}` | affinity for Minio deployment |
| minio.enabled | bool | `true` |  |
| minio.image.pullPolicy | string | `"IfNotPresent"` |  |
| minio.image.repository | string | `"minio/minio"` | Docker image for Minio deployment |
| minio.image.tag | string | `"RELEASE.2020-12-16T05-05-17Z"` |  |
| minio.nodeSelector | object | `{}` | nodeSelector for Minio deployment |
| minio.podAnnotations | object | `{}` | Annotations for Minio pods |
| minio.replicaCount | int | `1` | Replicas count for Minio deployment |
| minio.resources | object | `{"limits":{"cpu":"200m","memory":"512Mi"},"requests":{"cpu":"10m","memory":"128Mi"}}` | Default resources requests and limits for Minio deployment |
| minio.service | object | `{"annotations":{},"type":"ClusterIP"}` | Service settings for Minio |
| minio.tolerations | list | `[]` | tolerations for Minio deployment |
| postgres.affinity | object | `{}` | affinity for Postgres deployment |
| postgres.enabled | bool | `true` |  |
| postgres.image.pullPolicy | string | `"IfNotPresent"` |  |
| postgres.image.repository | string | `"postgres"` | Docker image for Postgres deployment |
| postgres.image.tag | string | `"10.16-alpine"` |  |
| postgres.nodeSelector | object | `{}` | nodeSelector for Postgres deployment |
| postgres.podAnnotations | object | `{}` | Annotations for Postgres pods |
| postgres.replicaCount | int | `1` | Replicas count for Postgres deployment |
| postgres.resources | object | `{"limits":{"cpu":"1000m","memory":"512Mi"},"requests":{"cpu":"10m","memory":"128Mi"}}` | Default resources requests and limits for Postgres deployment |
| postgres.service | object | `{"annotations":{},"type":"ClusterIP"}` | Service settings for Postgres |
| postgres.tolerations | list | `[]` | tolerations for Postgres deployment |
| pytorchoperator.enabled | bool | `false` |  |
| redis.enabled | bool | `false` |  |
| sagemaker.enabled | bool | `false` |  |
| sparkoperator.enabled | bool | `false` |  |
| storage.bucketName | string | `"my-s3-bucket"` | bucketName defines the storage bucket flyte will use. Required for all types except for sandbox. |
| storage.custom | object | `{}` | GCP project ID. Required for storage type gcs. projectId: -- Settings for storage type custom. See https://github:com/graymeta/stow for supported storage providers/settings. |
| storage.gcs | string | `nil` | settings for storage type gcs |
| storage.s3 | object | `{"region":"us-east-1"}` | settings for storage type s3 |
| storage.type | string | `"sandbox"` | Sets the storage type. Supported values are sandbox, s3, gcs and custom. |
| tf_operator.enabled | bool | `false` |  |
| webhook.enabled | bool | `true` | enable or disable secrets webhook |
| webhook.service | object | `{"annotations":{"projectcontour.io/upstream-protocol.h2c":"grpc"},"type":"ClusterIP"}` | Service settings for the webhook |
| webhook.serviceAccount | object | `{"annotations":{},"create":true,"imagePullSecrets":{}}` | Configuration for service accounts for the webhook |
| webhook.serviceAccount.annotations | object | `{}` | Annotations for ServiceAccount attached to the webhook |
| webhook.serviceAccount.create | bool | `true` | Should a service account be created for the webhook |
| webhook.serviceAccount.imagePullSecrets | object | `{}` | ImapgePullSecrets to automatically assign to the service account |
| workflow_notifications | object | `{"config":{},"enabled":false}` | **Optional Component** Workflow notifications module is an optional dependency. Flyte uses cloud native pub-sub systems to notify users of various events in their workflows |
| workflow_scheduler.config | object | `{}` |  |
| workflow_scheduler.enabled | bool | `false` |  |
