{{/* Expand the chart name. */}}
{{- define "flyte-core.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Create the resource name prefix. */}}
{{- define "flyte-core.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else if .Values.nameOverride -}}
{{- .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/* Prefix a resource name only when an override explicitly configures one. */}}
{{- define "flyte-core.resourceName" -}}
{{- $prefix := include "flyte-core.fullname" .root -}}
{{- if $prefix -}}
{{- printf "%s-%s" $prefix .name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- .name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "flyte-core.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "flyte-core.labels" -}}
helm.sh/chart: {{ include "flyte-core.chart" . }}
app.kubernetes.io/name: {{ include "flyte-core.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.commonLabels }}
{{ tpl (toYaml .) $ }}
{{- end }}
{{- end -}}

{{- define "flyte-core.componentName" -}}
{{- include "flyte-core.resourceName" (dict "root" .root "name" .component) -}}
{{- end -}}

{{- define "flyte-core.componentServiceHost" -}}
{{- printf "%s.%s" (include "flyte-core.componentName" .) .root.Release.Namespace -}}
{{- end -}}

{{- define "flyte-core.componentPort" -}}
{{- $configuration := index .root.Values.configuration .component -}}
{{- $configuration.server.port -}}
{{- end -}}

{{- define "flyte-core.componentServiceURL" -}}
{{- printf "http://%s:%s" (include "flyte-core.componentServiceHost" .) (include "flyte-core.componentPort" .) -}}
{{- end -}}

{{- define "flyte-core.componentSelectorLabels" -}}
app.kubernetes.io/name: {{ include "flyte-core.name" .root }}
app.kubernetes.io/instance: {{ .root.Release.Name }}
app.kubernetes.io/component: {{ .component }}
{{- end -}}

{{- define "flyte-core.componentLabels" -}}
{{ include "flyte-core.labels" .root }}
app.kubernetes.io/component: {{ .component }}
{{- end -}}

{{- define "flyte-core.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "flyte-core.resourceName" (dict "root" . "name" "service-account")) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{- define "flyte-core.clusterRoleName" -}}
{{- include "flyte-core.resourceName" (dict "root" . "name" "cluster-role") -}}
{{- end -}}

{{- define "flyte-core.componentConfigMapName" -}}
{{- include "flyte-core.componentName" . -}}
{{- end -}}

{{- define "flyte-core.configSecretName" -}}
{{- include "flyte-core.resourceName" (dict "root" . "name" "config-secret") -}}
{{- end -}}

{{- define "flyte-core.externalConfiguration" -}}
{{- or .Values.configuration.externalConfigMap .Values.configuration.externalSecretRef -}}
{{- end -}}

{{- define "flyte-core.executorCacheServiceName" -}}
{{- printf "%s-cache" (include "flyte-core.componentName" (dict "root" . "component" "executor")) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "flyte-core.executorCacheServiceHost" -}}
{{- printf "%s.%s" (include "flyte-core.executorCacheServiceName" .) .Release.Namespace -}}
{{- end -}}

{{- define "flyte-core.executorCacheServiceURL" -}}
{{- printf "http://%s:%v" (include "flyte-core.executorCacheServiceHost" .) .Values.configuration.webhook.cacheInvalidationPort -}}
{{- end -}}

{{- define "flyte-core.webhookSecretName" -}}
{{- include "flyte-core.resourceName" (dict "root" . "name" "webhook-certs") -}}
{{- end -}}

{{- define "flyte-core.consoleName" -}}
{{- include "flyte-core.resourceName" (dict "root" . "name" "console") -}}
{{- end -}}

{{- define "flyte-core.config.logger" -}}
logger:
  level: {{ .Values.configuration.logging.level }}
  show-source: {{ .Values.configuration.logging.showSource }}
{{- end -}}

{{- define "flyte-core.config.database" -}}
{{- $database := deepCopy .Values.configuration.database }}
{{- if $database.postgres }}{{- $_ := unset $database.postgres "password" }}{{- end }}
database:
  {{- tpl (toYaml $database) . | nindent 2 }}
{{- end -}}

{{- define "flyte-core.config.storage" -}}
{{- with .Values.configuration.storage }}
storage:
  type: stow
  stow:
    {{- if eq "s3" .provider }}
    kind: s3
    config:
      region: {{ required "configuration.storage.providerConfig.s3.region is required" .providerConfig.s3.region }}
      disable_ssl: {{ .providerConfig.s3.disableSSL }}
      v2_signing: {{ .providerConfig.s3.v2Signing }}
      {{- with .providerConfig.s3.endpoint }}
      endpoint: {{ tpl . $ | quote }}
      {{- end }}
      {{- if eq "iam" .providerConfig.s3.authType }}
      auth_type: iam
      {{- else if eq "accesskey" .providerConfig.s3.authType }}
      auth_type: accesskey
      {{- else }}
      {{- fail "configuration.storage.providerConfig.s3.authType must be iam or accesskey" }}
      {{- end }}
    {{- else if eq "gcs" .provider }}
    kind: google
    config:
      json: ""
      project_id: {{ required "configuration.storage.providerConfig.gcs.project is required" .providerConfig.gcs.project }}
      scopes: https://www.googleapis.com/auth/cloud-platform
    {{- else if eq "azure" .provider }}
    kind: azure
    config:
      account: {{ required "configuration.storage.providerConfig.azure.account is required" .providerConfig.azure.account }}
      {{- with .providerConfig.azure.configDomainSuffix }}
      configDomainSuffix: {{ . | quote }}
      {{- end }}
      configUploadConcurrency: {{ .providerConfig.azure.configUploadConcurrency }}
    {{- else }}
    {{- fail "configuration.storage.provider must be s3, gcs, or azure" }}
    {{- end }}
  container: {{ required "configuration.storage.metadataContainer is required" .metadataContainer | quote }}
{{- end }}
{{- end -}}

{{/* Public API routes and the split services that serve them. */}}
{{- define "flyte-core.ingress.apiRoutes" -}}
- component: runs
  paths:
    - /flyteidl2.workflow.RunService
    - /flyteidl2.task.TaskService
    - /flyteidl2.trigger.TriggerService
    - /flyteidl2.project.ProjectService
    - /flyteidl2.auth.IdentityService
- component: actions
  paths:
    - /flyteidl2.actions.ActionsService
- component: events
  paths:
    - /flyteidl2.workflow.EventsProxyService
- component: cache
  paths:
    - /flyteidl2.cacheservice.v2.CacheService
- component: dataproxy
  paths:
    - /flyteidl2.dataproxy.DataProxyService
    - /flyteidl2.cluster.ClusterService
    - /flyteidl2.workflow.TranslatorService
- component: secret
  paths:
    - /flyteidl2.secret.SecretService
- component: app
  paths:
    - /flyteidl2.app.AppService
    - /flyteidl2.app.AppLogsService
{{- end -}}

{{/* Unauthenticated OAuth discovery routes served by the Runs service. */}}
{{- define "flyte-core.ingress.wellknownRoutes" -}}
- component: runs
  paths:
    - /.well-known/oauth-authorization-server
    - /flyteidl2.auth.AuthMetadataService
{{- end -}}
