# Flyte Single Deployment Helm Chart

Chart for the single Flyte executable style of deployment

![Version: v0.1.10](https://img.shields.io/badge/Version-v0.1.10-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.16.0](https://img.shields.io/badge/AppVersion-1.16.0-informational?style=flat-square)

## Management of Helm Values that are Secret
Flyte as a platform over the course of its normal operation requires access to multiple resources that may be gated by the secrets. For example, the database might have a password, and if OAuth2 is enabled, there are multiple secrets that the control and data plane portion of Flyte will need access to. All these secrets that the Flyte binary relies on all ultimately come through in the form of files. There are two types of files that can contain secrets.

### How Secrets are Loaded

The secrets that the flyte binary relies on all ultimately come through in through files.  There are two types of files that can contain secrets.
* Configuration Files to the Flyte Executable
  These are a series of yaml files that get loaded via command line args to the executable.  Configuration for the same key (for maps, not lists) can be split across different files.  At run time, flyte's configuration system will merge the files to form one cohesive view of the configuration.  By default these are mounted under `/etc/flyte/config.d/` in the container. For example, if you were to set the database password via a configuration file, it would look like.
  ```
    database:
      postgres:
        username: postgres
        password: mypassword
  ```
  This would get picked up by flyte's configuration logic that looks for the [db password](https://github.com/flyteorg/flyteadmin/blob/53282fe979b628447e72d0e3f8fd0e2b9235d929/pkg/runtime/interfaces/application_configuration.go#L45).

* Raw Secret Files
  These are files that contain just a straight value that gets read as a string value and should be the actual password you're looking for.  By default these are mounted under `/etc/secrets/` in the container.

### Secrets with this Chart
If you're trying to get a secret into the flyte executable's set of configuration files, there's two ways of going about this.

* Set the value as plaintext in your Helm values file. This is the least secure, default option and if you're reading this you probably want the two options below. For example, the database password above is hooked up to the `chart/flyte-binary` Helm chart's value under
  ```
  configuration:
  database:
    password: "mypassword"
  ```
  If you do this, the Helm chart will create a K8s Secret whose value mimics the configuration structure that the flyte binary expects:
  ```
  012-database-secrets.yaml: |
    database:
      postgres:
        password: "mypassword"
  ```
  At run time, the binary will read this in and merge it into the configuration that ultimately gets used.

* Set the value separately as a secret.  The advantage of doing this is that you don't have to have a secret sitting as plaintext in your Helm values file. Create an external secret containing info such as DB password, S3 access/secret key, client secret hash, etc.

  ```
  $ cat <<EOF | kubectl apply -f -
  apiVersion: v1
  kind: Secret
  metadata:
    name: flyte-binary-inline-config-secret
    namespace: flyte
  type: Opaque
  stringData:
    202-database-secrets.yaml: |
      database:
        postgres:
          password: <DB_PASSWORD>
    203-storage-secrets.yaml: |
      storage:
        stow:
          config:
            access_key_id: <S3_ACCESS_KEY>
            secret_key: <S3_SECRET_KEY>
    204-auth-secrets.yaml: |
      auth:
        appAuth:
          selfAuthServer:
            staticClients:
              flytepropeller:
                client_secret: <CLIENT_SECRET_HASH>
  EOF
  ```
  Then reference the newly created secret in `.Values.configuration.inlineSecretRef` in `values.yaml` by setting:

  ```
  configuration:
    inlineSecretRef: flyte-binary-inline-config-secret
  ```

  This option basically mirrors the first option, except that you're creating the secret out of band (using TF or some other process).  This style of specifying secrets mimics the "inline" method for specifying additional configuration, hence the name.

* For the secrets that always expect a path (currently just the OIDC secret and the client credentials secret that Flyte uses to talk to itself), you can store using K8s secrets as follows. First create an external secret containing the secret values:

  ```
  $ cat <<EOF | kubectl apply -f -
  apiVersion: v1
  kind: Secret
  metadata:
    name: flyte-binary-client-secrets-external-secret
    namespace: flyte
  type: Opaque
  stringData:
    client_secret: <INTERNAL_CLIENT_SECRET>
    oidc_client_secret: <OIDC_CLIENT_SECRET>
  EOF
  ```

  Then reference the newly created secret in `.Values.configuration.auth.clientSecretsExternalSecretRef` in `values.yaml` as follows:
  ```
  configuration:
    auth:
      clientSecretsExternalSecretRef: flyte-binary-client-secrets-external-secret
  ```

  The `flyte-binary` Helm chart is set up such that these get mounted directly under the `/etc/secrets/` folder.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| clusterResourceTemplates.annotations | object | `{}` |  |
| clusterResourceTemplates.externalConfigMap | string | `""` |  |
| clusterResourceTemplates.inline | object | `{}` |  |
| clusterResourceTemplates.inlineConfigMap | string | `""` |  |
| clusterResourceTemplates.labels | object | `{}` |  |
| commonAnnotations | object | `{}` |  |
| commonLabels | object | `{}` |  |
| configuration.agentService | object | `{}` |  |
| configuration.annotations | object | `{}` |  |
| configuration.auth.authorizedUris | list | `[]` |  |
| configuration.auth.clientSecretsExternalSecretRef | string | `""` |  |
| configuration.auth.enableAuthServer | bool | `true` |  |
| configuration.auth.enabled | bool | `false` |  |
| configuration.auth.flyteClient.audience | string | `""` |  |
| configuration.auth.flyteClient.clientId | string | `"flytectl"` |  |
| configuration.auth.flyteClient.redirectUri | string | `"http://localhost:53593/callback"` |  |
| configuration.auth.flyteClient.scopes[0] | string | `"all"` |  |
| configuration.auth.internal.clientId | string | `"flytepropeller"` |  |
| configuration.auth.internal.clientSecret | string | `""` |  |
| configuration.auth.internal.clientSecretHash | string | `""` |  |
| configuration.auth.oidc.baseUrl | string | `""` |  |
| configuration.auth.oidc.clientId | string | `""` |  |
| configuration.auth.oidc.clientSecret | string | `""` |  |
| configuration.co-pilot.image.repository | string | `"cr.flyte.org/flyteorg/flytecopilot"` |  |
| configuration.co-pilot.image.tag | string | `"v0.0.30"` |  |
| configuration.database.dbname | string | `"flyte"` |  |
| configuration.database.host | string | `"127.0.0.1"` |  |
| configuration.database.options | string | `"sslmode=disable"` |  |
| configuration.database.password | string | `""` |  |
| configuration.database.passwordPath | string | `""` |  |
| configuration.database.port | int | `5432` |  |
| configuration.database.username | string | `"postgres"` |  |
| configuration.externalConfigMap | string | `""` |  |
| configuration.externalSecretRef | string | `""` |  |
| configuration.inline | object | `{}` |  |
| configuration.inlineConfigMap | string | `""` |  |
| configuration.inlineSecretRef | string | `""` |  |
| configuration.labels | object | `{}` |  |
| configuration.logging.level | int | `1` |  |
| configuration.logging.plugins.cloudwatch.enabled | bool | `false` |  |
| configuration.logging.plugins.cloudwatch.templateUri | string | `""` |  |
| configuration.logging.plugins.custom | list | `[]` |  |
| configuration.logging.plugins.kubernetes.enabled | bool | `false` |  |
| configuration.logging.plugins.kubernetes.templateUri | string | `""` |  |
| configuration.logging.plugins.stackdriver.enabled | bool | `false` |  |
| configuration.logging.plugins.stackdriver.templateUri | string | `""` |  |
| configuration.storage.metadataContainer | string | `"my-organization-flyte-container"` |  |
| configuration.storage.provider | string | `"s3"` |  |
| configuration.storage.providerConfig.gcs.project | string | `"my-organization-gcp-project"` |  |
| configuration.storage.providerConfig.s3.accessKey | string | `""` |  |
| configuration.storage.providerConfig.s3.authType | string | `"iam"` |  |
| configuration.storage.providerConfig.s3.disableSSL | bool | `false` |  |
| configuration.storage.providerConfig.s3.endpoint | string | `""` |  |
| configuration.storage.providerConfig.s3.region | string | `"us-east-1"` |  |
| configuration.storage.providerConfig.s3.secretKey | string | `""` |  |
| configuration.storage.providerConfig.s3.v2Signing | bool | `false` |  |
| configuration.storage.userDataContainer | string | `"my-organization-flyte-container"` |  |
| deployment.annotations | object | `{}` |  |
| deployment.args | list | `[]` |  |
| deployment.command | list | `[]` |  |
| deployment.extraEnvVars | list | `[]` |  |
| deployment.extraEnvVarsConfigMap | string | `""` |  |
| deployment.extraEnvVarsSecret | string | `""` |  |
| deployment.extraPodSpec | object | `{}` |  |
| deployment.extraVolumeMounts | list | `[]` |  |
| deployment.extraVolumes | list | `[]` |  |
| deployment.genAdminAuthSecret.args | list | `[]` |  |
| deployment.genAdminAuthSecret.command | list | `[]` |  |
| deployment.image.pullPolicy | string | `"IfNotPresent"` |  |
| deployment.image.repository | string | `"cr.flyte.org/flyteorg/flyte-binary"` |  |
| deployment.image.tag | string | `"latest"` |  |
| deployment.initContainers | list | `[]` |  |
| deployment.labels | object | `{}` |  |
| deployment.lifecycleHooks | object | `{}` |  |
| deployment.livenessProbe | object | `{}` |  |
| deployment.podAnnotations | object | `{}` |  |
| deployment.podLabels | object | `{}` |  |
| deployment.podSecurityContext.enabled | bool | `false` |  |
| deployment.podSecurityContext.fsGroup | int | `65534` |  |
| deployment.podSecurityContext.runAsGroup | int | `65534` |  |
| deployment.podSecurityContext.runAsUser | int | `65534` |  |
| deployment.readinessProbe | object | `{}` |  |
| deployment.sidecars | list | `[]` |  |
| deployment.startupProbe | object | `{}` |  |
| deployment.waitForDB.args | list | `[]` |  |
| deployment.waitForDB.command | list | `[]` |  |
| deployment.waitForDB.image.pullPolicy | string | `"IfNotPresent"` |  |
| deployment.waitForDB.image.repository | string | `"postgres"` |  |
| deployment.waitForDB.image.tag | string | `"15-alpine"` |  |
| enabled_plugins.tasks | object | `{"task-plugins":{"default-for-task-types":{"container":"container","container_array":"k8s-array","sidecar":"sidecar"},"enabled-plugins":["container","sidecar","k8s-array"]}}` | Tasks specific configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#GetConfig) |
| enabled_plugins.tasks.task-plugins | object | `{"default-for-task-types":{"container":"container","container_array":"k8s-array","sidecar":"sidecar"},"enabled-plugins":["container","sidecar","k8s-array"]}` | Plugins configuration, [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#TaskPluginConfig) |
| enabled_plugins.tasks.task-plugins.enabled-plugins | list | `["container","sidecar","k8s-array"]` | [Enabled Plugins](https://pkg.go.dev/github.com/lyft/flyteplugins/go/tasks/config#Config). Enable sagemaker*, athena if you install the backend plugins |
| flyte-core-components.admin.disableClusterResourceManager | bool | `false` |  |
| flyte-core-components.admin.disableScheduler | bool | `false` |  |
| flyte-core-components.admin.disabled | bool | `false` |  |
| flyte-core-components.admin.seedProjects[0] | string | `"flytesnacks"` |  |
| flyte-core-components.dataCatalog.disabled | bool | `false` |  |
| flyte-core-components.propeller.disableWebhook | bool | `false` |  |
| flyte-core-components.propeller.disabled | bool | `false` |  |
| flyteagent.deployment.annotations | object | `{}` |  |
| flyteagent.deployment.args | list | `[]` |  |
| flyteagent.deployment.command | list | `[]` |  |
| flyteagent.deployment.extraEnvVars | list | `[]` |  |
| flyteagent.deployment.extraEnvVarsConfigMap | string | `""` |  |
| flyteagent.deployment.extraEnvVarsSecret | string | `""` |  |
| flyteagent.deployment.extraPodSpec | object | `{}` |  |
| flyteagent.deployment.extraVolumeMounts | list | `[]` |  |
| flyteagent.deployment.extraVolumes | list | `[]` |  |
| flyteagent.deployment.image.pullPolicy | string | `"IfNotPresent"` |  |
| flyteagent.deployment.image.repository | string | `"ghcr.io/flyteorg/flyteagent"` |  |
| flyteagent.deployment.image.tag | string | `"1.6.2b1"` |  |
| flyteagent.deployment.initContainers | list | `[]` |  |
| flyteagent.deployment.labels | object | `{}` |  |
| flyteagent.deployment.lifecycleHooks | object | `{}` |  |
| flyteagent.deployment.livenessProbe | object | `{}` |  |
| flyteagent.deployment.podAnnotations | object | `{}` |  |
| flyteagent.deployment.podLabels | object | `{}` |  |
| flyteagent.deployment.podSecurityContext.enabled | bool | `false` |  |
| flyteagent.deployment.podSecurityContext.fsGroup | int | `65534` |  |
| flyteagent.deployment.podSecurityContext.runAsGroup | int | `65534` |  |
| flyteagent.deployment.podSecurityContext.runAsUser | int | `65534` |  |
| flyteagent.deployment.readinessProbe | object | `{}` |  |
| flyteagent.deployment.replicas | int | `1` |  |
| flyteagent.deployment.sidecars | list | `[]` |  |
| flyteagent.deployment.startupProbe | object | `{}` |  |
| flyteagent.enable | bool | `false` |  |
| flyteagent.service.annotations | object | `{}` |  |
| flyteagent.service.clusterIP | string | `""` |  |
| flyteagent.service.externalTrafficPolicy | string | `"Cluster"` |  |
| flyteagent.service.extraPorts | list | `[]` |  |
| flyteagent.service.labels | object | `{}` |  |
| flyteagent.service.loadBalancerIP | string | `""` |  |
| flyteagent.service.loadBalancerSourceRanges | list | `[]` |  |
| flyteagent.service.nodePort | string | `""` |  |
| flyteagent.service.port | string | `""` |  |
| flyteagent.service.type | string | `"ClusterIP"` |  |
| flyteagent.serviceAccount.annotations | object | `{}` |  |
| flyteagent.serviceAccount.create | bool | `true` |  |
| flyteagent.serviceAccount.imagePullSecrets | list | `[]` |  |
| flyteagent.serviceAccount.labels | object | `{}` |  |
| flyteagent.serviceAccount.name | string | `""` |  |
| fullnameOverride | string | `""` |  |
| ingress.commonAnnotations | object | `{}` |  |
| ingress.create | bool | `false` |  |
| ingress.grpcAnnotations | object | `{}` |  |
| ingress.grpcExtraPaths.append | list | `[]` |  |
| ingress.grpcExtraPaths.prepend | list | `[]` |  |
| ingress.host | string | `""` |  |
| ingress.httpAnnotations | object | `{}` |  |
| ingress.httpExtraPaths.append | list | `[]` |  |
| ingress.httpExtraPaths.prepend | list | `[]` |  |
| ingress.labels | object | `{}` |  |
| nameOverride | string | `""` |  |
| rbac.annotations | object | `{}` |  |
| rbac.create | bool | `true` |  |
| rbac.extraRules | list | `[]` |  |
| rbac.labels | object | `{}` |  |
| service.clusterIP | string | `""` |  |
| service.commonAnnotations | object | `{}` |  |
| service.externalTrafficPolicy | string | `"Cluster"` |  |
| service.extraPorts | list | `[]` |  |
| service.grpcAnnotations | object | `{}` |  |
| service.httpAnnotations | object | `{}` |  |
| service.labels | object | `{}` |  |
| service.loadBalancerIP | string | `""` |  |
| service.loadBalancerSourceRanges | list | `[]` |  |
| service.nodePorts.grpc | string | `""` |  |
| service.nodePorts.http | string | `""` |  |
| service.ports.grpc | string | `""` |  |
| service.ports.http | string | `""` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.imagePullSecrets | list | `[]` |  |
| serviceAccount.labels | object | `{}` |  |
| serviceAccount.name | string | `""` |  |

