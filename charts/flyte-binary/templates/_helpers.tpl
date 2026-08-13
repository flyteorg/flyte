{{/*
Expand the name of the chart.
*/}}
{{- define "flyte-binary.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "flyte-binary.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "flyte-binary.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Base labels
*/}}
{{- define "flyte-binary.baseLabels" -}}
app.kubernetes.io/name: {{ include "flyte-binary.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "flyte-binary.labels" -}}
helm.sh/chart: {{ include "flyte-binary.chart" . }}
{{ include "flyte-binary.baseLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "flyte-binary.selectorLabels" -}}
{{ include "flyte-binary.baseLabels" . }}
app.kubernetes.io/component: flyte-binary
{{- end }}

{{/*
Console fully qualified name
*/}}
{{- define "flyte-binary.console.fullname" -}}
{{- printf "%s-console" (include "flyte-binary.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Console service name
*/}}
{{- define "flyte-binary.console.serviceName" -}}
{{- include "flyte-binary.console.fullname" . -}}
{{- end -}}

{{/*
Console selector labels
*/}}
{{- define "flyte-binary.console.selectorLabels" -}}
{{ include "flyte-binary.baseLabels" . }}
app.kubernetes.io/component: console
{{- end -}}

{{/*
Console common labels
*/}}
{{- define "flyte-binary.console.labels" -}}
helm.sh/chart: {{ include "flyte-binary.chart" . }}
{{ include "flyte-binary.console.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "flyte-binary.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "flyte-binary.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Flag to use external configuration.
*/}}
{{- define "flyte-binary.configuration.externalConfiguration" -}}
{{- or .Values.configuration.externalConfigMap .Values.configuration.externalSecretRef -}}
{{- end -}}

{{/*
The Secret an operator supplied for co-pilot, if any. Its own key rather than a field on
configuration.storage, since it configures co-pilot rather than the deployment's storage,
and the section name means it needs the index dance.
*/}}
{{- define "flyte-binary.configuration.copilotStorageSecretRef" -}}
{{- (index .Values.configuration "co-pilot").storageSecretRef -}}
{{- end -}}

{{/*
Fail on value combinations the chart would otherwise ignore in silence. Called from
deployment.yaml, which renders whatever else is switched off.
*/}}
{{- define "flyte-binary.validateValues" -}}
{{- if and (include "flyte-binary.configuration.copilotStorageSecretRef" .) (include "flyte-binary.configuration.externalConfiguration" .) -}}
{{- fail "configuration.co-pilot.storageSecretRef has no effect while configuration.externalConfigMap or configuration.externalSecretRef is set: the chart renders no ConfigMap, and that is what would carry plugins.k8s.co-pilot.storage-config-secret-name. Set that key in your own configuration instead, then unset storageSecretRef." -}}
{{- end -}}
{{- end -}}

{{/*
Get the Flyte configuration ConfigMap name.
*/}}
{{- define "flyte-binary.configuration.configMapName" -}}
{{- printf "%s-config" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{/*
Get the Flyte configuration Secret name.
*/}}
{{- define "flyte-binary.configuration.configSecretName" -}}
{{- printf "%s-config-secret" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{/*
Get the name of the Secret holding co-pilot's storage configuration.
*/}}
{{- define "flyte-binary.configuration.copilotStorageSecretName" -}}
{{- printf "%s-copilot-storage-config-secret" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{/*
The stow storage block, shared by the ConfigMap (003-storage.yaml) and co-pilot's Secret so
the two cannot drift. Call with (dict "root" $ "withCredentials" bool); only the Secret sets
withCredentials, since the ConfigMap keeps them in 013-storage-secrets.yaml instead.
*/}}
{{- define "flyte-binary.configuration.storageBlock" -}}
{{- $root := .root -}}
{{- $withCredentials := .withCredentials -}}
{{- with $root.Values.configuration.storage -}}
storage:
  type: stow
  stow:
    {{- if eq "s3" .provider }}
    {{- with .providerConfig.s3 }}
    kind: s3
    config:
      region: {{ required "Region required for S3 storage provider" .region }}
      disable_ssl: {{ .disableSSL }}
      v2_signing: {{ .v2Signing }}
      {{- if .endpoint }}
      endpoint: {{ tpl .endpoint $root }}
      {{- end }}
      {{- if eq "iam" .authType }}
      auth_type: iam
      {{- else if eq "accesskey" .authType }}
      auth_type: accesskey
      {{- if $withCredentials }}
      access_key_id: {{ required "Access key required for S3 storage provider" .accessKey | quote }}
      secret_key: {{ required "secretKey required for co-pilot storage configuration" .secretKey | quote }}
      {{- end }}
      {{- else }}
      {{- printf "Invalid value for S3 storage provider authentication type. Expected one of (iam, accesskey), but got: %s" .authType | fail }}
      {{- end }}
    {{- end }}
    {{- else if eq "azure" .provider }}
    {{- with .providerConfig.azure }}
    kind: azure
    config:
      account: {{ .account }}
      {{- if .key }}
      key: {{ .key }}
      {{- end }}
      {{- if .configDomainSuffix }}
      configDomainSuffix: {{ .configDomainSuffix }}
      {{- end }}
      {{- if .configUploadConcurrency }}
      configUploadConcurrency: {{ .configUploadConcurrency }}
      {{- end }}
    {{- end }}
    {{- else if eq "gcs" .provider }}
    kind: google
    config:
      json: ""
      project_id: {{ required "GCP project required for GCS storage provider" .providerConfig.gcs.project }}
      scopes: https://www.googleapis.com/auth/cloud-platform
    {{- else }}
    {{- printf "Invalid value for storage provider. Expected one of (s3, azure, gcs), but got: %s" .provider | fail }}
    {{- end }}
  container: {{ required "Metadata container required" .metadataContainer }}
{{- end }}
{{- end -}}

{{/*
Whether co-pilot reads its storage configuration from a Secret rather than from its command
line. The Secret exists to keep credentials out of every task's pod spec, so it is rendered
only when there is a credential to keep out of them. Ambient auth (S3 authType=iam, GCS
workload identity, Azure without a key) puts nothing sensitive on the command line, and a
Secret carrying a region and an endpoint would be machinery protecting nothing.

Rendering the Secret and pointing co-pilot at it are both gated on this and must stay in
lockstep: naming a Secret that is never rendered leaves every task pod stuck in
ContainerCreating.

configuration.inline counts, whatever it holds. The chart cannot see whether an operator
put a session token or a GCS service-account key in there, and config-secret.yaml merges
it into the Secret, so err towards the Secret rather than towards the command line.

S3 with only secretKeyPath is the one credential-bearing case that stays on the command
line: the path names a file living solely in this deployment's container, so the chart
cannot read it to render a Secret. Those deployments supply their own via
co-pilot.storageSecretRef.
*/}}
{{- define "flyte-binary.configuration.copilotStorageFromSecret" -}}
{{- $root := . -}}
{{- with .Values.configuration.storage -}}
{{- if include "flyte-binary.configuration.copilotStorageSecretRef" $root -}}
true
{{- else if and (eq "s3" .provider) (eq "accesskey" .providerConfig.s3.authType) .providerConfig.s3.secretKey -}}
true
{{- else if and (eq "azure" .provider) .providerConfig.azure.key -}}
true
{{- else if and $root.Values.configuration.inline (hasKey $root.Values.configuration.inline "storage") -}}
true
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Get the Flyte logging configuration.
*/}}
{{- define "flyte-binary.configuration.logging.plugins" -}}
{{- with .Values.configuration.logging.plugins -}}
kubernetes-enabled: {{ .kubernetes.enabled }}
{{- if .kubernetes.enabled }}
kubernetes-template-uri: {{ required "Template URI required for Kubernetes logging plugin" .kubernetes.templateUri }}
{{- end }}
cloudwatch-enabled: {{ .cloudwatch.enabled }}
{{- if .cloudwatch.enabled }}
cloudwatch-template-uri: {{ required "Template URI required for CloudWatch logging plugin" .cloudwatch.templateUri }}
{{- end }}
stackdriver-enabled: {{ .stackdriver.enabled }}
{{- if .stackdriver.enabled }}
stackdriver-template-uri: {{ required "Template URI required for stackdriver logging plugin" .stackdriver.templateUri }}
{{- end }}
{{- if .custom }}
templates: {{- toYaml .custom | nindent 2 -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Get the Flyte cluster resource templates ConfigMap name.
*/}}
{{- define "flyte-binary.clusterResourceTemplates.configMapName" -}}
{{- printf "%s-cluster-resource-templates" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{/*
Get the Flyte HTTP service name
*/}}
{{- define "flyte-binary.service.http.name" -}}
{{- printf "%s-http" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{/*
Get the Flyte service HTTP port.
*/}}
{{- define "flyte-binary.service.http.port" -}}
{{- default 8090 .Values.service.ports.http -}}
{{- end -}}

{{/*
Get the Flyte API paths for ingress.
*/}}
{{- define "flyte-binary.ingress.apiPaths" -}}
- /flyteidl2.workflow.RunService
- /flyteidl2.workflow.RunService/*
- /flyteidl2.workflow.RunLogsService
- /flyteidl2.workflow.RunLogsService/*
- /flyteidl2.task.TaskService
- /flyteidl2.task.TaskService/*
- /flyteidl2.workflow.TranslatorService
- /flyteidl2.workflow.TranslatorService/*
- /flyteidl2.dataproxy.DataProxyService
- /flyteidl2.dataproxy.DataProxyService/*
- /flyteidl2.cluster.ClusterService
- /flyteidl2.cluster.ClusterService/*
- /flyteidl2.secret.SecretService
- /flyteidl2.secret.SecretService/*
- /flyteidl2.project.ProjectService
- /flyteidl2.project.ProjectService/*
- /flyteidl2.app.AppService
- /flyteidl2.app.AppService/*
- /flyteidl2.app.AppLogsService
- /flyteidl2.app.AppLogsService/*
- /flyteidl2.trigger.TriggerService
- /flyteidl2.trigger.TriggerService/*
- /flyteidl2.auth.IdentityService
- /flyteidl2.auth.IdentityService/*
- /flyteidl2.settings.SettingsService
- /flyteidl2.settings.SettingsService/*
{{- end -}}

{{/*
Get the Flyte auth-discovery paths for ingress. These are unauthenticated:
clients must reach them before they hold a token (OAuth server metadata and the
auth metadata service). IdentityService and SettingsService require auth and live
in apiPaths instead.
*/}}
{{- define "flyte-binary.ingress.wellknownPaths" -}}
- /.well-known/oauth-authorization-server
- /flyteidl2.auth.AuthMetadataService
- /flyteidl2.auth.AuthMetadataService/*
{{- end -}}

{{/*
Get the Flyte webhook service name.
*/}}
{{- define "flyte-binary.webhook.serviceName" -}}
{{- printf "%s-webhook" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{- define "flyte-binary.webhook.headlessServiceName" -}}
{{- printf "%s-webhook-headless" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{/*
Get the Flyte webhook secret name.
*/}}
{{- define "flyte-binary.webhook.secretName" -}}
{{- printf "%s-webhook-secret" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{/*
Get the Flyte ClusterRole name.
*/}}
{{- define "flyte-binary.rbac.clusterRoleName" -}}
{{- printf "%s-cluster-role" (include "flyte-binary.fullname" .) -}}
{{- end -}}
