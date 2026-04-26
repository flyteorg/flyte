{{/*
Expand the name of the chart.
*/}}
{{- define "flyte-core.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "flyte-core.fullname" -}}
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
Chart label value.
*/}}
{{- define "flyte-core.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels applied to every resource.
*/}}
{{- define "flyte-core.labels" -}}
helm.sh/chart: {{ include "flyte-core.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: flyte
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Component labels — call with (dict "ctx" $ "component" "runs").
*/}}
{{- define "flyte-core.componentLabels" -}}
{{ include "flyte-core.labels" .ctx }}
app.kubernetes.io/name: {{ include "flyte-core.name" .ctx }}
app.kubernetes.io/component: {{ .component }}
{{- end }}

{{/*
Selector labels for a component — call with (dict "ctx" $ "component" "runs").
*/}}
{{- define "flyte-core.selectorLabels" -}}
app.kubernetes.io/name: {{ include "flyte-core.name" .ctx }}
app.kubernetes.io/instance: {{ .ctx.Release.Name }}
app.kubernetes.io/component: {{ .component }}
{{- end }}

{{/*
ServiceAccount name.
*/}}
{{- define "flyte-core.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "flyte-core.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Component fullname — call with (dict "ctx" $ "component" "runs").
*/}}
{{- define "flyte-core.componentFullname" -}}
{{- printf "%s-%s" (include "flyte-core.fullname" .ctx) .component | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Per-component ConfigMap name — call with (dict "ctx" $ "component" "runs").
*/}}
{{- define "flyte-core.componentConfigMapName" -}}
{{- printf "%s-%s-config" (include "flyte-core.fullname" .ctx) .component | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Database Secret name.
*/}}
{{- define "flyte-core.dbSecretName" -}}
{{- printf "%s-db-password" (include "flyte-core.fullname" .) }}
{{- end }}

{{/*
Webhook service name.
*/}}
{{- define "flyte-core.webhook.serviceName" -}}
{{- printf "%s-webhook" (include "flyte-core.fullname" .) }}
{{- end }}

{{/*
ClusterRole name.
*/}}
{{- define "flyte-core.clusterRoleName" -}}
{{- printf "%s-cluster-role" (include "flyte-core.fullname" .) }}
{{- end }}

{{/*
Storage stow config block — reused across component configmaps.
*/}}
{{- define "flyte-core.storageConfig" -}}
storage:
  type: stow
  stow:
    {{- if eq .Values.storage.provider "s3" }}
    {{- with .Values.storage.providerConfig.s3 }}
    kind: s3
    config:
      region: {{ .region }}
      disable_ssl: {{ .disableSSL }}
      v2_signing: {{ .v2Signing }}
      {{- if .endpoint }}
      endpoint: {{ .endpoint }}
      {{- end }}
      auth_type: {{ .authType }}
    {{- end }}
    {{- else if eq .Values.storage.provider "gcs" }}
    kind: google
    config:
      json: ""
      project_id: {{ .Values.storage.providerConfig.gcs.project }}
      scopes: https://www.googleapis.com/auth/cloud-platform
    {{- else if eq .Values.storage.provider "azure" }}
    {{- with .Values.storage.providerConfig.azure }}
    kind: azure
    config:
      account: {{ .account }}
      {{- if .key }}
      key: {{ .key }}
      {{- end }}
    {{- end }}
    {{- end }}
  container: {{ .Values.storage.metadataContainer }}
{{- end }}

{{/*
Database config block — reused by runs and cache_service.
*/}}
{{- define "flyte-core.databaseConfig" -}}
database:
  postgres:
    username: {{ .Values.database.username }}
    host: {{ tpl .Values.database.host . }}
    port: {{ .Values.database.port }}
    dbname: {{ .Values.database.dbname }}
    options: {{ .Values.database.options | quote }}
    {{- if .Values.database.password }}
    passwordPath: /etc/flyte/db-password/password
    {{- else if .Values.database.passwordPath }}
    passwordPath: {{ .Values.database.passwordPath }}
    {{- end }}
{{- end }}
