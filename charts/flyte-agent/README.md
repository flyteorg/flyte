# flyteagent

![Version: v0.1.10](https://img.shields.io/badge/Version-v0.1.10-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)

A Helm chart for Flyte agent

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| agentSecret.name | string | `""` | Specify name of K8s Secret. Leave it empty if you don't need this Secret |
| agentSecret.secretData | object | `{"data":{"username":"User"}}` | Specify your Secret (with sensitive data) or pseudo-manifest (without sensitive data). See https://github.com/godaddy/kubernetes-external-secrets |
| commonAnnotations | object | `{}` |  |
| commonLabels | object | `{}` |  |
| flyteagent.flyteagent.additionalContainers | list | `[]` | Appends additional containers to the deployment spec. May include template values. |
| flyteagent.flyteagent.additionalVolumeMounts | list | `[]` | Appends additional volume mounts to the main container's spec. May include template values. |
| flyteagent.flyteagent.additionalVolumes | list | `[]` | Appends additional volumes to the deployment spec. May include template values. |
| flyteagent.flyteagent.affinity | object | `{}` | affinity for flyteagent deployment |
| flyteagent.flyteagent.configPath | string | `"/etc/flyteagent/config/*.yaml"` | Default regex string for searching configuration files |
| flyteagent.flyteagent.enabled | bool | `true` |  |
| flyteagent.flyteagent.extraArgs | object | `{}` | Appends extra command line arguments to the main command |
| flyteagent.flyteagent.image.pullPolicy | string | `"IfNotPresent"` | Docker image pull policy |
| flyteagent.flyteagent.image.repository | string | `"ghcr.io/flyteorg/flyteagent"` | Docker image for flyteagent deployment |
| flyteagent.flyteagent.image.tag | string | `"1.8.3"` | Docker image tag |
| flyteagent.flyteagent.nodeSelector | object | `{}` | nodeSelector for flyteagent deployment |
| flyteagent.flyteagent.podAnnotations | object | `{}` | Annotations for flyteagent pods |
| flyteagent.flyteagent.ports.containerPort | int | `8000` |  |
| flyteagent.flyteagent.ports.name | string | `"agent-grpc"` |  |
| flyteagent.flyteagent.priorityClassName | string | `""` | Sets priorityClassName for datacatalog pod(s). |
| flyteagent.flyteagent.replicaCount | int | `1` | Replicas count for flyteagent deployment |
| flyteagent.flyteagent.resources | object | `{"limits":{"cpu":"500m","ephemeral-storage":"200Mi","memory":"200Mi"},"requests":{"cpu":"500m","ephemeral-storage":"200Mi","memory":"200Mi"}}` | Default resources requests and limits for flyteagent deployment |
| flyteagent.flyteagent.service | object | `{"annotations":{"projectcontour.io/upstream-protocol.h2c":"grpc"},"type":"ClusterIP"}` | Service settings for flyteagent |
| flyteagent.flyteagent.serviceAccount | object | `{"annotations":{},"create":true,"imagePullSecrets":[]}` | Configuration for service accounts for flyteagent |
| flyteagent.flyteagent.serviceAccount.annotations | object | `{}` | Annotations for ServiceAccount attached to flyteagent pods |
| flyteagent.flyteagent.serviceAccount.create | bool | `true` | Should a service account be created for flyteagent |
| flyteagent.flyteagent.serviceAccount.imagePullSecrets | list | `[]` | ImagePullSecrets to automatically assign to the service account |
| flyteagent.flyteagent.tolerations | list | `[]` | tolerations for flyteagent deployment |
| fullnameOverride | string | `""` |  |
| nameOverride | string | `""` |  |
