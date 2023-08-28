# flyteagent

![Version: v0.1.10](https://img.shields.io/badge/Version-v0.1.10-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)

A Helm chart for Flyte agent

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| additionalContainers | list | `[]` | Appends additional containers to the deployment spec. May include template values. |
| additionalVolumeMounts | list | `[]` | Appends additional volume mounts to the main container's spec. May include template values. |
| additionalVolumes | list | `[]` | Appends additional volumes to the deployment spec. May include template values. |
| affinity | object | `{}` | affinity for flyteagent deployment |
| agentSecret.secretData | object | `{"data":{"username":"User"}}` | Specify your Secret (with sensitive data) or pseudo-manifest (without sensitive data). |
| commonAnnotations | object | `{}` |  |
| commonLabels | object | `{}` |  |
| configPath | string | `"/etc/flyteagent/config/*.yaml"` | Default regex string for searching configuration files |
| extraArgs | object | `{}` | Appends extra command line arguments to the main command |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` | Docker image pull policy |
| image.repository | string | `"ghcr.io/flyteorg/flyteagent"` | Docker image for flyteagent deployment |
| image.tag | string | `"1.8.3"` | Docker image tag |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` | nodeSelector for flyteagent deployment |
| podAnnotations | object | `{}` | Annotations for flyteagent pods |
| ports.containerPort | int | `8000` |  |
| ports.name | string | `"agent-grpc"` |  |
| priorityClassName | string | `""` | Sets priorityClassName for datacatalog pod(s). |
| replicaCount | int | `1` | Replicas count for flyteagent deployment |
| resources | object | `{"limits":{"cpu":"500m","ephemeral-storage":"200Mi","memory":"200Mi"},"requests":{"cpu":"500m","ephemeral-storage":"200Mi","memory":"200Mi"}}` | Default resources requests and limits for flyteagent deployment |
| service | object | `{"annotations":{"projectcontour.io/upstream-protocol.h2c":"grpc"},"type":"ClusterIP"}` | Service settings for flyteagent |
| serviceAccount | object | `{"annotations":{},"create":true,"imagePullSecrets":[]}` | Configuration for service accounts for flyteagent |
| serviceAccount.annotations | object | `{}` | Annotations for ServiceAccount attached to flyteagent pods |
| serviceAccount.create | bool | `true` | Should a service account be created for flyteagent |
| serviceAccount.imagePullSecrets | list | `[]` | ImagePullSecrets to automatically assign to the service account |
| tolerations | list | `[]` | tolerations for flyteagent deployment |
