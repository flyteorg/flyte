# flyte

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.8.0](https://img.shields.io/badge/AppVersion-0.8.0-informational?style=flat-square)

A Helm chart for Flyte

### SANDBOX INSTALLATION:
- [Install helm 3](https://helm.sh/docs/intro/install/) 
- Fetch chart dependencies `helm dep up`
- Install Flyte sandbox:

```bash
cd helm
helm install -n flyte -f values-sandbox.yaml --create-namespace flyte .
```

Customize your installation by changing settings in `values-sandbox.yaml`.
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
- Get flyte host `minikube service contour -n heptio-contour --url`. And then visit `http://<HOST>/console`


### CONFIGURATION NOTES:
- The docker images, their tags and other default parameters are configured in `values.yaml` file.
- Each Flyte installation type should have separate `values-*.yaml` file: for sandbox, EKS and etc. The configuration in `values.yaml` and the choosen config `values-*.yaml` are merged when generating the deployment manifest.
- The configuration in `values-sandbox.yaml` is ready for installation in minikube. But `values-eks.yaml` should be edited before installation: s3 bucket, RDS hosts, iam roles, secrets and etc need to be modified.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| common.databaseSecret.name | string | `""` | Specify name of K8s Secret which contains Database password. Leave it empty if you don't need this Secret  |
| common.databaseSecret.secretManifest | object | `{}` | Specify your Secret (with sensitive data) or pseudo-manifest (without sensitive data). See https://github.com/godaddy/kubernetes-external-secrets |
| common.flyteNamespaceTemplate.enabled | bool | `true` |  |
| common.ingress.annotations | object | `{}` |  |
| common.ingress.enabled | bool | `true` |  |
| configmap.aa_namespace | string | `"apiVersion: v1\nkind: Namespace\nmetadata:\n  name: {{ namespace }}\nspec:\n  finalizers:\n  - kubernetes\n"` | Template for namespaces resources |
| configmap.ad_spark_role | string | `"apiVersion: rbac.authorization.k8s.io/v1beta1\nkind: Role\nmetadata:\n  name: spark-role\n  namespace: {{ namespace }}\nrules:\n- apiGroups:\n  - \"\"\n  resources:\n  - pods\n  verbs:\n  - '*'\n- apiGroups:\n  - \"\"\n  resources:\n  - services\n  verbs:\n  - '*'\n- apiGroups:\n  - \"\"\n  resources:\n  - configmaps\n  verbs:\n  - '*'\n"` | Template for spark role |
| configmap.admin.admin.endpoint | string | `"flyteadmin:81"` |  |
| configmap.admin.admin.insecure | bool | `true` |  |
| configmap.admin.event.capacity | int | `1000` |  |
| configmap.admin.event.rate | int | `500` |  |
| configmap.admin.event.type | string | `"admin"` |  |
| configmap.ae_spark_service_account | string | `"apiVersion: v1\nkind: ServiceAccount\nmetadata:\n  name: spark\n  namespace: {{ namespace }}\n"` | Template for spark SA |
| configmap.af_spark_role_binding | string | `"apiVersion: rbac.authorization.k8s.io/v1beta1\nkind: RoleBinding\nmetadata:\n  name: spark-role-binding\n  namespace: {{ namespace }} \nroleRef:\n  apiGroup: rbac.authorization.k8s.io\n  kind: Role\n  name: spark-role\nsubjects:\n- kind: ServiceAccount\n  name: spark\n  namespace: {{ namespace }}\n"` | Template for spark role binding |
| configmap.catalog.catalog-cache.endpoint | string | `"datacatalog:89"` |  |
| configmap.catalog.catalog-cache.insecure | bool | `true` |  |
| configmap.catalog.catalog-cache.type | string | `"datacatalog"` |  |
| configmap.catalog_cache.plugins.catalogCache.reader.maxItems | int | `10000` |  |
| configmap.catalog_cache.plugins.catalogCache.writer.maxItems | int | `10000` |  |
| configmap.cluster_resources.cluster_resources.customData.development[0].projectQuotaCpu.value | string | `"4"` |  |
| configmap.cluster_resources.cluster_resources.customData.development[1].projectQuotaMemory.value | string | `"3000Mi"` |  |
| configmap.cluster_resources.cluster_resources.customData.production[0].projectQuotaCpu.value | string | `"5"` |  |
| configmap.cluster_resources.cluster_resources.customData.production[1].projectQuotaMemory.value | string | `"4000Mi"` |  |
| configmap.cluster_resources.cluster_resources.customData.staging[0].projectQuotaCpu.value | string | `"2"` |  |
| configmap.cluster_resources.cluster_resources.customData.staging[1].projectQuotaMemory.value | string | `"3000Mi"` |  |
| configmap.cluster_resources.cluster_resources.refresh | string | `"5m"` |  |
| configmap.cluster_resources.cluster_resources.templatePath | string | `"/etc/flyte/clusterresource/templates"` |  |
| configmap.console.BASE_URL | string | `"/console"` |  |
| configmap.console.CONFIG_DIR | string | `"/etc/flyte/config"` |  |
| configmap.copilot.plugins.k8s.co-pilot.image | string | `"docker.io/lyft/flytecopilot:v0.3.35"` |  |
| configmap.copilot.plugins.k8s.co-pilot.name | string | `"flyte-copilot-"` |  |
| configmap.copilot.plugins.k8s.co-pilot.start-timeout | string | `"30s"` |  |
| configmap.core.propeller.downstream-eval-duration | string | `"30s"` |  |
| configmap.core.propeller.enable-admin-launcher | bool | `true` |  |
| configmap.core.propeller.leader-election.enabled | bool | `true` |  |
| configmap.core.propeller.leader-election.lease-duration | string | `"15s"` |  |
| configmap.core.propeller.leader-election.lock-config-map.name | string | `"propeller-leader"` |  |
| configmap.core.propeller.leader-election.lock-config-map.namespace | string | `"flyte"` |  |
| configmap.core.propeller.leader-election.renew-deadline | string | `"10s"` |  |
| configmap.core.propeller.leader-election.retry-period | string | `"2s"` |  |
| configmap.core.propeller.limit-namespace | string | `"all"` |  |
| configmap.core.propeller.max-workflow-retries | int | `30` |  |
| configmap.core.propeller.metadata-prefix | string | `"metadata/propeller"` |  |
| configmap.core.propeller.metrics-prefix | string | `"flyte"` |  |
| configmap.core.propeller.prof-port | int | `10254` |  |
| configmap.core.propeller.queue.batch-size | int | `-1` |  |
| configmap.core.propeller.queue.batching-interval | string | `"2s"` |  |
| configmap.core.propeller.queue.queue.capacity | int | `100` |  |
| configmap.core.propeller.queue.queue.rate | int | `10` |  |
| configmap.core.propeller.queue.queue.type | string | `"bucket"` |  |
| configmap.core.propeller.queue.sub-queue.capacity | int | `100` |  |
| configmap.core.propeller.queue.sub-queue.rate | int | `10` |  |
| configmap.core.propeller.queue.sub-queue.type | string | `"bucket"` |  |
| configmap.core.propeller.queue.type | string | `"batch"` |  |
| configmap.core.propeller.rawoutput-prefix | string | `"s3://my-s3-bucket/"` |  |
| configmap.core.propeller.workers | int | `4` |  |
| configmap.core.propeller.workflow-reeval-duration | string | `"30s"` |  |
| configmap.datacatalogServer.datacatalog.metrics-scope | string | `"datacatalog"` |  |
| configmap.datacatalogServer.datacatalog.profiler-port | int | `10254` |  |
| configmap.datacatalogServer.datacatalog.storage-prefix | string | `"metadata/datacatalog"` |  |
| configmap.db.database | object | `{}` |  |
| configmap.domain.domains[0].id | string | `"development"` |  |
| configmap.domain.domains[0].name | string | `"development"` |  |
| configmap.domain.domains[1].id | string | `"staging"` |  |
| configmap.domain.domains[1].name | string | `"staging"` |  |
| configmap.domain.domains[2].id | string | `"production"` |  |
| configmap.domain.domains[2].name | string | `"production"` |  |
| configmap.enabled_plugins.tasks.max-plugin-phase-versions | int | `1000000` |  |
| configmap.enabled_plugins.tasks.task-plugins.enabled-plugins | list | `[]` |  |
| configmap.k8s.plugins.k8s.default-cpus | string | `"100m"` |  |
| configmap.k8s.plugins.k8s.default-env-vars | list | `[]` |  |
| configmap.k8s.plugins.k8s.default-memory | string | `"100Mi"` |  |
| configmap.logger.logger.level | int | `4` |  |
| configmap.logger.logger.show-source | bool | `true` |  |
| configmap.qubole.plugins.qubole.quboleTokenKey | string | `"FLYTE_QUBOLE_CLIENT_TOKEN"` |  |
| configmap.remote_data.remoteData.region | string | `"us-east-1"` |  |
| configmap.remote_data.remoteData.scheme | string | `"local"` |  |
| configmap.remote_data.remoteData.signedUrls.durationMinutes | int | `3` |  |
| configmap.resource_manager.propeller.resourcemanager.redis.hostKey | string | `"mypassword"` |  |
| configmap.resource_manager.propeller.resourcemanager.redis.hostPath | string | `"redis-resource-manager:6379"` |  |
| configmap.resource_manager.propeller.resourcemanager.resourceMaxQuota | int | `10000` |  |
| configmap.resource_manager.propeller.resourcemanager.type | string | `"redis"` |  |
| configmap.server.flyteadmin.metadataStoragePrefix[0] | string | `"metadata"` |  |
| configmap.server.flyteadmin.metadataStoragePrefix[1] | string | `"admin"` |  |
| configmap.server.flyteadmin.metricsScope | string | `"flyte:"` |  |
| configmap.server.flyteadmin.profilerPort | int | `10254` |  |
| configmap.server.flyteadmin.roleNameKey | string | `"iam.amazonaws.com/role"` |  |
| configmap.server.flyteadmin.testing.host | string | `"http://flyteadmin"` |  |
| configmap.server.server.grpcPort | int | `8089` |  |
| configmap.server.server.httpPort | int | `8088` |  |
| configmap.server.server.security.allowCors | bool | `true` |  |
| configmap.server.server.security.allowedHeaders[0] | string | `"Content-Type"` |  |
| configmap.server.server.security.allowedOrigins[0] | string | `"*"` |  |
| configmap.server.server.security.secure | bool | `false` |  |
| configmap.server.server.security.useAuth | bool | `false` |  |
| configmap.spark.plugins.spark.spark-config-default | list | `[{"spark.hadoop.mapreduce.fileoutputcommitter.algorithm.version":"2"},{"spark.kubernetes.allocation.batch.size":"50"},{"spark.hadoop.fs.s3a.acl.default":"BucketOwnerFullControl"},{"spark.hadoop.fs.s3n.impl":"org.apache.hadoop.fs.s3a.S3AFileSystem"},{"spark.hadoop.fs.AbstractFileSystem.s3n.impl":"org.apache.hadoop.fs.s3a.S3A"},{"spark.hadoop.fs.s3.impl":"org.apache.hadoop.fs.s3a.S3AFileSystem"},{"spark.hadoop.fs.AbstractFileSystem.s3.impl":"org.apache.hadoop.fs.s3a.S3A"},{"spark.hadoop.fs.s3a.impl":"org.apache.hadoop.fs.s3a.S3AFileSystem"},{"spark.hadoop.fs.AbstractFileSystem.s3a.impl":"org.apache.hadoop.fs.s3a.S3A"},{"spark.hadoop.fs.s3a.multipart.threshold":"536870912"},{"spark.blacklist.enabled":"true"},{"spark.blacklist.timeout":"5m"},{"spark.task.maxfailures":"8"}]` | Spark default configuration |
| configmap.storage.storage.connection.auth-type | string | `"accesskey"` |  |
| configmap.storage.storage.connection.region | string | `"us-east-1"` |  |
| configmap.storage.storage.container | string | `"my-s3-bucket"` |  |
| configmap.storage.storage.type | string | `"minio"` |  |
| configmap.task_logs.plugins.logs.kubernetes-enabled | bool | `false` |  |
| configmap.task_resource_defaults.task_resources.defaults.cpu | string | `"100m"` |  |
| configmap.task_resource_defaults.task_resources.defaults.memory | string | `"1000Mi"` |  |
| configmap.task_resource_defaults.task_resources.defaults.storage | string | `"5Mi"` |  |
| configmap.task_resource_defaults.task_resources.limits.cpu | string | `"15000m"` |  |
| configmap.task_resource_defaults.task_resources.limits.memory | string | `"100Gi"` |  |
| configmap.task_resource_defaults.task_resources.limits.storage | string | `"400Gi"` |  |
| contour.affinity | object | `{}` | affinity for Contour deployment |
| contour.contour.image.pullPolicy | string | `"IfNotPresent"` |  |
| contour.contour.image.repository | string | `"gcr.io/heptio-images/contour"` | Docker image for Contour |
| contour.contour.image.tag | string | `"v0.6.1"` |  |
| contour.contour.resources | object | `{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"10m","memory":"50Mi"}}` | Default resources requests and limits for Contour |
| contour.enabled | bool | `true` |  |
| contour.envoy.image.pullPolicy | string | `"IfNotPresent"` |  |
| contour.envoy.image.repository | string | `"docker.io/envoyproxy/envoy-alpine"` | Docker image for Envoy |
| contour.envoy.image.tag | string | `"v1.6.0"` |  |
| contour.envoy.resources | object | `{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"10m","memory":"50Mi"}}` | Default resources requests and limits for Envoy |
| contour.nodeSelector | object | `{}` | nodeSelector for Contour deployment |
| contour.podAnnotations | object | `{}` | Annotations for Contour pods |
| contour.replicaCount | int | `1` | Replicas count for Contour deployment |
| contour.service | object | `{"annotations":{},"ports":[{"nodePort":30081,"port":80,"protocol":"TCP"}],"type":"NodePort"}` | Service settings for Contour |
| contour.serviceAccountAnnotations | object | `{}` | Annotations for ServiceAccount attached to Contour pods |
| contour.tolerations | list | `[]` | tolerations for Contour deployment |
| datacatalog.affinity | object | `{}` | affinity for Datacatalog deployment |
| datacatalog.configPath | string | `"/etc/datacatalog/config/*.yaml"` | Default regex string for searching configuration files |
| datacatalog.image.pullPolicy | string | `"IfNotPresent"` |  |
| datacatalog.image.repository | string | `"docker.io/lyft/datacatalog"` | Docker image for Datacatalog deployment |
| datacatalog.image.tag | string | `"v0.2.2"` |  |
| datacatalog.nodeSelector | object | `{}` | nodeSelector for Datacatalog deployment |
| datacatalog.podAnnotations | object | `{}` | Annotations for Datacatalog pods |
| datacatalog.replicaCount | int | `1` | Replicas count for Datacatalog deployment |
| datacatalog.resources | object | `{"limits":{"cpu":"500m","ephemeral-storage":"100Mi","memory":"500Mi"},"requests":{"cpu":"10m","ephemeral-storage":"50Mi","memory":"50Mi"}}` | Default resources requests and limits for Datacatalog deployment |
| datacatalog.service | object | `{"annotations":{"contour.heptio.com/upstream-protocol.h2c":"grpc"},"type":"NodePort"}` | Service settings for Datacatalog |
| datacatalog.serviceAccountAnnotations | object | `{}` | Annotations for ServiceAccount attached to Datacatalog pods |
| datacatalog.tolerations | list | `[]` | tolerations for Datacatalog deployment |
| flyteadmin.affinity | object | `{}` | affinity for Flyteadmin deployment |
| flyteadmin.configPath | string | `"/etc/flyte/config/*.yaml"` | Default regex string for searching configuration files |
| flyteadmin.image.pullPolicy | string | `"IfNotPresent"` |  |
| flyteadmin.image.repository | string | `"docker.io/lyft/flyteadmin"` | Docker image for Flyteadmin deployment |
| flyteadmin.image.tag | string | `"v0.3.7"` |  |
| flyteadmin.nodeSelector | object | `{}` | nodeSelector for Flyteadmin deployment |
| flyteadmin.podAnnotations | object | `{}` | Annotations for Flyteadmin pods |
| flyteadmin.replicaCount | int | `1` | Replicas count for Flyteadmin deployment |
| flyteadmin.resources | object | `{"limits":{"cpu":"250m","ephemeral-storage":"100Mi","memory":"500Mi"},"requests":{"cpu":"10m","ephemeral-storage":"50Mi","memory":"50Mi"}}` | Default resources requests and limits for Flyteadmin deployment |
| flyteadmin.service | object | `{"annotations":{"contour.heptio.com/upstream-protocol.h2c":"grpc"},"type":"ClusterIP"}` | Service settings for Flyteadmin |
| flyteadmin.serviceAccountAnnotations | object | `{}` | Annotations for ServiceAccount attached to Flyteadmin pods |
| flyteadmin.tolerations | list | `[]` | tolerations for Flyteadmin deployment |
| flyteconsole.affinity | object | `{}` | affinity for Flyteconsole deployment |
| flyteconsole.image.pullPolicy | string | `"IfNotPresent"` |  |
| flyteconsole.image.repository | string | `"docker.io/lyft/flyteconsole"` | Docker image for Flyteconsole deployment |
| flyteconsole.image.tag | string | `"v0.12.1"` |  |
| flyteconsole.nodeSelector | object | `{}` | nodeSelector for Flyteconsole deployment |
| flyteconsole.podAnnotations | object | `{}` | Annotations for Flyteconsole pods |
| flyteconsole.replicaCount | int | `1` | Replicas count for Flyteconsole deployment |
| flyteconsole.resources | object | `{"limits":{"cpu":"500m","memory":"250Mi"},"requests":{"cpu":"10m","memory":"50Mi"}}` | Default resources requests and limits for Flyteconsole deployment |
| flyteconsole.service | object | `{"annotations":{},"type":"ClusterIP"}` | Service settings for Flyteconsole |
| flyteconsole.tolerations | list | `[]` | tolerations for Flyteconsole deployment |
| flytepropeller.affinity | object | `{}` | affinity for Flytepropeller deployment |
| flytepropeller.configPath | string | `"/etc/flyte/config/*.yaml"` | Default regex string for searching configuration files |
| flytepropeller.image.pullPolicy | string | `"IfNotPresent"` |  |
| flytepropeller.image.repository | string | `"docker.io/lyft/flytepropeller"` | Docker image for Flytepropeller deployment |
| flytepropeller.image.tag | string | `"v0.4.2"` |  |
| flytepropeller.nodeSelector | object | `{}` | nodeSelector for Flytepropeller deployment |
| flytepropeller.podAnnotations | object | `{}` | Annotations for Flytepropeller pods |
| flytepropeller.replicaCount | int | `1` | Replicas count for Flytepropeller deployment |
| flytepropeller.resources | object | `{"limits":{"cpu":"200m","ephemeral-storage":"100Mi","memory":"200Mi"},"requests":{"cpu":"10m","ephemeral-storage":"50Mi","memory":"50Mi"}}` | Default resources requests and limits for Flytepropeller deployment |
| flytepropeller.serviceAccountAnnotations | object | `{}` | Annotations for ServiceAccount attached to Flytepropeller pods |
| flytepropeller.tolerations | list | `[]` | tolerations for Flytepropeller deployment |
| minio.affinity | object | `{}` | affinity for Minio deployment |
| minio.enabled | bool | `true` |  |
| minio.image.pullPolicy | string | `"IfNotPresent"` |  |
| minio.image.repository | string | `"minio/minio"` | Docker image for Minio deployment |
| minio.image.tag | string | `"RELEASE.2019-06-04T01-15-58Z"` |  |
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
| postgres.image.tag | string | `"10.1"` |  |
| postgres.nodeSelector | object | `{}` | nodeSelector for Postgres deployment |
| postgres.podAnnotations | object | `{}` | Annotations for Postgres pods |
| postgres.replicaCount | int | `1` | Replicas count for Postgres deployment |
| postgres.resources | object | `{"limits":{"cpu":"1000m","memory":"512Mi"},"requests":{"cpu":"10m","memory":"128Mi"}}` | Default resources requests and limits for Postgres deployment |
| postgres.service | object | `{"annotations":{},"type":"ClusterIP"}` | Service settings for Postgres |
| postgres.tolerations | list | `[]` | tolerations for Postgres deployment |
| pytorchoperator.affinity | object | `{}` | affinity for Pytorchoperator deployment |
| pytorchoperator.enabled | bool | `true` |  |
| pytorchoperator.image.pullPolicy | string | `"IfNotPresent"` |  |
| pytorchoperator.image.repository | string | `"gcr.io/kubeflow-images-public/pytorch-operator"` | Docker image for Pytorchoperator |
| pytorchoperator.image.tag | string | `"v1.0.0-g047cf0f"` |  |
| pytorchoperator.nodeSelector | object | `{}` | nodeSelector for Pytorchoperator deployment |
| pytorchoperator.podAnnotations | object | `{}` | Annotations for Pytorchoperator pods |
| pytorchoperator.replicaCount | int | `1` | Replicas count for Pytorchoperator deployment |
| pytorchoperator.resources | object | `{"limits":{"cpu":"500m","memory":"1000M"},"requests":{"cpu":"10m","memory":"50M"}}` | Default resources requests and limits for Pytorchoperator |
| pytorchoperator.service | object | `{"annotations":{},"type":"ClusterIP"}` | Service settings for Pytorchoperator |
| pytorchoperator.serviceAccountAnnotations | object | `{}` | Annotations for ServiceAccount attached to Pytorchoperator pods |
| pytorchoperator.tolerations | list | `[]` | tolerations for Pytorchoperator deployment |
| redis.affinity | object | `{}` | affinity for Redis Statefulset |
| redis.enabled | bool | `true` |  |
| redis.image.pullPolicy | string | `"IfNotPresent"` |  |
| redis.image.repository | string | `"docker.io/bitnami/redis"` | Docker image for Redis Statefulset |
| redis.image.tag | string | `"4.0.2-r1"` |  |
| redis.nodeSelector | object | `{}` | nodeSelector for Redis Statefulset |
| redis.podAnnotations | object | `{}` | Annotations for Redis pods |
| redis.replicaCount | int | `1` | Replicas count for Redis Statefulset |
| redis.resources | object | `{"limits":{"cpu":"1000m","memory":"1Gi"},"requests":{"cpu":"10m","memory":"50Mi"}}` | Default resources requests and limits for Redis Statefulset |
| redis.service | object | `{"annotations":{},"type":"ClusterIP"}` | Service settings for Redis |
| redis.tolerations | list | `[]` | tolerations for Redis Statefulset |
| sparkoperator.affinity | object | `{}` | affinity for Sparkoperator deployment |
| sparkoperator.enabled | bool | `true` |  |
| sparkoperator.image.pullPolicy | string | `"IfNotPresent"` |  |
| sparkoperator.image.repository | string | `"gcr.io/spark-operator/spark-operator"` | Docker image for Sparkoperator |
| sparkoperator.image.tag | string | `"v2.4.0-v1beta1-0.9.0"` |  |
| sparkoperator.nodeSelector | object | `{}` | nodeSelector for Sparkoperator deployment |
| sparkoperator.podAnnotations | object | `{}` | Annotations for Sparkoperator pods |
| sparkoperator.replicaCount | int | `1` | Replicas count for Sparkoperator deployment |
| sparkoperator.resources | object | `{"limits":{"cpu":"1000m","memory":"500M"},"requests":{"cpu":"10m","memory":"50M"}}` | Default resources requests and limits for Sparkoperator |
| sparkoperator.service | object | `{"annotations":{},"type":"ClusterIP"}` | Service settings for Contour |
| sparkoperator.serviceAccountAnnotations | object | `{}` | Annotations for ServiceAccount attached to Sparkoperator pods |
| sparkoperator.tolerations | list | `[]` | tolerations for Sparkoperator deployment |
