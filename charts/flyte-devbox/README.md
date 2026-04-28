# flyte-devbox

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.16.1](https://img.shields.io/badge/AppVersion-1.16.1-informational?style=flat-square)

A Helm chart for the Flyte local demo cluster

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../flyte-binary | flyte-binary | v0.2.0 |
| https://charts.bitnami.com/bitnami | minio | 12.6.7 |
| https://twuni.github.io/docker-registry.helm | docker-registry | 2.2.2 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| docker-registry.enabled | bool | `true` |  |
| docker-registry.fullnameOverride | string | `"docker-registry"` |  |
| docker-registry.image.pullPolicy | string | `"Never"` |  |
| docker-registry.image.tag | string | `"sandbox"` |  |
| docker-registry.persistence.enabled | bool | `false` |  |
| docker-registry.secrets.haSharedSecret | string | `"flytesandboxsecret"` |  |
| docker-registry.service.nodePort | int | `30000` |  |
| docker-registry.service.type | string | `"NodePort"` |  |
| flyte-binary.clusterResourceTemplates.inlineConfigMap | string | `"{{ include \"flyte-devbox.clusterResourceTemplates.inlineConfigMap\" . }}"` |  |
| flyte-binary.configuration.database.host | string | `"postgresql"` |  |
| flyte-binary.configuration.database.password | string | `"postgres"` |  |
| flyte-binary.configuration.inline.plugins.k8s.default-env-vars[0].FLYTE_AWS_ENDPOINT | string | `"http://minio.{{ .Release.Namespace }}:9000"` |  |
| flyte-binary.configuration.inline.plugins.k8s.default-env-vars[1].FLYTE_AWS_ACCESS_KEY_ID | string | `"minio"` |  |
| flyte-binary.configuration.inline.plugins.k8s.default-env-vars[2].FLYTE_AWS_SECRET_ACCESS_KEY | string | `"miniostorage"` |  |
| flyte-binary.configuration.inline.plugins.k8s.default-env-vars[3]._U_EP_OVERRIDE | string | `"flyte-binary-http.{{ .Release.Namespace }}:8090"` |  |
| flyte-binary.configuration.inline.plugins.k8s.default-env-vars[4]._U_INSECURE | string | `"true"` |  |
| flyte-binary.configuration.inline.plugins.k8s.default-env-vars[5]._U_USE_ACTIONS | string | `"1"` |  |
| flyte-binary.configuration.inline.runs.database.postgres.dbName | string | `"runs"` |  |
| flyte-binary.configuration.inline.runs.database.postgres.host | string | `"postgresql.{{ .Release.Namespace }}"` |  |
| flyte-binary.configuration.inline.runs.database.postgres.password | string | `"postgres"` |  |
| flyte-binary.configuration.inline.runs.database.postgres.port | int | `5432` |  |
| flyte-binary.configuration.inline.runs.database.postgres.user | string | `"postgres"` |  |
| flyte-binary.configuration.inline.storage.signedURL.stowConfigOverride.endpoint | string | `"http://localhost:30002"` |  |
| flyte-binary.configuration.inline.task_resources.defaults.cpu | string | `"500m"` |  |
| flyte-binary.configuration.inline.task_resources.defaults.ephemeralStorage | int | `0` |  |
| flyte-binary.configuration.inline.task_resources.defaults.gpu | int | `0` |  |
| flyte-binary.configuration.inline.task_resources.defaults.memory | string | `"1Gi"` |  |
| flyte-binary.configuration.inline.task_resources.limits.cpu | int | `0` |  |
| flyte-binary.configuration.inline.task_resources.limits.ephemeralStorage | int | `0` |  |
| flyte-binary.configuration.inline.task_resources.limits.gpu | int | `0` |  |
| flyte-binary.configuration.inline.task_resources.limits.memory | int | `0` |  |
| flyte-binary.configuration.inlineConfigMap | string | `"{{ include \"flyte-devbox.configuration.inlineConfigMap\" . }}"` |  |
| flyte-binary.configuration.logging.level | int | `5` |  |
| flyte-binary.configuration.storage.metadataContainer | string | `"flyte-data"` |  |
| flyte-binary.configuration.storage.provider | string | `"s3"` |  |
| flyte-binary.configuration.storage.providerConfig.s3.accessKey | string | `"minio"` |  |
| flyte-binary.configuration.storage.providerConfig.s3.authType | string | `"accesskey"` |  |
| flyte-binary.configuration.storage.providerConfig.s3.disableSSL | bool | `true` |  |
| flyte-binary.configuration.storage.providerConfig.s3.endpoint | string | `"http://minio.{{ .Release.Namespace }}:9000"` |  |
| flyte-binary.configuration.storage.providerConfig.s3.secretKey | string | `"miniostorage"` |  |
| flyte-binary.configuration.storage.providerConfig.s3.v2Signing | bool | `true` |  |
| flyte-binary.configuration.storage.userDataContainer | string | `"flyte-data"` |  |
| flyte-binary.deployment.image.pullPolicy | string | `"Never"` |  |
| flyte-binary.deployment.image.repository | string | `"flyte-binary-v2"` |  |
| flyte-binary.deployment.image.tag | string | `"sandbox"` |  |
| flyte-binary.deployment.livenessProbe.httpGet.path | string | `"/healthz"` |  |
| flyte-binary.deployment.livenessProbe.httpGet.port | string | `"http"` |  |
| flyte-binary.deployment.readinessProbe.httpGet.path | string | `"/readyz"` |  |
| flyte-binary.deployment.readinessProbe.httpGet.port | string | `"http"` |  |
| flyte-binary.deployment.readinessProbe.periodSeconds | int | `1` |  |
| flyte-binary.deployment.startupProbe.failureThreshold | int | `30` |  |
| flyte-binary.deployment.startupProbe.httpGet.path | string | `"/healthz"` |  |
| flyte-binary.deployment.startupProbe.httpGet.port | string | `"http"` |  |
| flyte-binary.deployment.startupProbe.periodSeconds | int | `1` |  |
| flyte-binary.deployment.waitForDB.image.pullPolicy | string | `"Never"` |  |
| flyte-binary.deployment.waitForDB.image.repository | string | `"bitnami/postgresql"` |  |
| flyte-binary.deployment.waitForDB.image.tag | string | `"sandbox"` |  |
| flyte-binary.enabled | bool | `true` |  |
| flyte-binary.fullnameOverride | string | `"flyte-binary"` |  |
| flyte-binary.rbac.extraRules[0].apiGroups[0] | string | `"*"` |  |
| flyte-binary.rbac.extraRules[0].resources[0] | string | `"*"` |  |
| flyte-binary.rbac.extraRules[0].verbs[0] | string | `"*"` |  |
| minio.auth.rootPassword | string | `"miniostorage"` |  |
| minio.auth.rootUser | string | `"minio"` |  |
| minio.defaultBuckets | string | `"flyte-data"` |  |
| minio.enabled | bool | `true` |  |
| minio.extraEnvVars[0].name | string | `"MINIO_BROWSER_REDIRECT_URL"` |  |
| minio.extraEnvVars[0].value | string | `"http://localhost:30080/minio"` |  |
| minio.fullnameOverride | string | `"minio"` |  |
| minio.image.pullPolicy | string | `"Never"` |  |
| minio.image.tag | string | `"sandbox"` |  |
| minio.persistence.enabled | bool | `true` |  |
| minio.persistence.existingClaim | string | `"{{ include \"flyte-devbox.persistence.minioVolumeName\" . }}"` |  |
| minio.service.nodePorts.api | int | `30002` |  |
| minio.service.type | string | `"NodePort"` |  |
| minio.volumePermissions.enabled | bool | `true` |  |
| minio.volumePermissions.image.pullPolicy | string | `"Never"` |  |
| minio.volumePermissions.image.tag | string | `"sandbox"` |  |
| postgresql.auth.postgresPassword | string | `"postgres"` |  |
| postgresql.enabled | bool | `true` |  |
| postgresql.fullnameOverride | string | `"postgresql"` |  |
| postgresql.image.pullPolicy | string | `"Never"` |  |
| postgresql.image.tag | string | `"sandbox"` |  |
| postgresql.primary.persistence.enabled | bool | `true` |  |
| postgresql.primary.persistence.existingClaim | string | `"{{ include \"flyte-devbox.persistence.dbVolumeName\" . }}"` |  |
| postgresql.primary.service.nodePorts.postgresql | int | `30001` |  |
| postgresql.primary.service.type | string | `"NodePort"` |  |
| postgresql.shmVolume.enabled | bool | `false` |  |
| postgresql.volumePermissions.enabled | bool | `true` |  |
| postgresql.volumePermissions.image.pullPolicy | string | `"Never"` |  |
| postgresql.volumePermissions.image.tag | string | `"sandbox"` |  |
| sandbox.console.enabled | bool | `true` |  |
| sandbox.console.image.pullPolicy | string | `"Never"` |  |
| sandbox.console.image.repository | string | `"ghcr.io/flyteorg/flyte-client-v2"` |  |
| sandbox.console.image.tag | string | `"latest"` |  |
| sandbox.dev | bool | `false` |  |

