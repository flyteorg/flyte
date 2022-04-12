# flyte-deps

![Version: v0.1.10](https://img.shields.io/badge/Version-v0.1.10-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)

A Helm chart for Flyte dependency

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | contour | 5.0.0 |
| https://googlecloudplatform.github.io/spark-on-k8s-operator | sparkoperator(spark-operator) | 1.1.15 |
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
| postgres.dbname | string | `"flyteadmin"` |  |
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
| redis.enabled | bool | `false` | - enable or disable Redis Statefulset installation |
| redoc.affinity | object | `{}` | affinity for Minio deployment |
| redoc.enabled | bool | `true` | - enable or disable Minio deployment installation |
| redoc.image.pullPolicy | string | `"IfNotPresent"` | Docker image pull policy |
| redoc.image.repository | string | `"docker.io/redocly/redoc"` | Docker image for Minio deployment |
| redoc.image.tag | string | `"latest"` | Docker image tag |
| redoc.nodeSelector | object | `{}` | nodeSelector for Minio deployment |
| redoc.podAnnotations | object | `{}` | Annotations for Minio pods |
| redoc.replicaCount | int | `1` | Replicas count for Minio deployment |
| redoc.resources | object | `{"limits":{"cpu":"200m","memory":"512Mi"},"requests":{"cpu":"10m","memory":"128Mi"}}` | Default resources requests and limits for Minio deployment |
| redoc.resources.limits | object | `{"cpu":"200m","memory":"512Mi"}` | Limits are the maximum set of resources needed for this pod |
| redoc.resources.requests | object | `{"cpu":"10m","memory":"128Mi"}` | Requests are the minimum set of resources needed for this pod |
| redoc.service | object | `{"type":"ClusterIP"}` | Service settings for Minio |
| redoc.tolerations | list | `[]` | tolerations for Minio deployment |
| sparkoperator | object | `{"enabled":false}` | Optional: Spark Plugin using the Spark Operator |
| sparkoperator.enabled | bool | `false` | - enable or disable Sparkoperator deployment installation |
