{{/*
Expand the name of the chart.
*/}}
{{- define "flyte-sandbox.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "flyte-sandbox.fullname" -}}
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
{{- define "flyte-sandbox.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "flyte-sandbox.labels" -}}
helm.sh/chart: {{ include "flyte-sandbox.chart" . }}
{{ include "flyte-sandbox.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "flyte-sandbox.selectorLabels" -}}
app.kubernetes.io/name: {{ include "flyte-sandbox.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "flyte-sandbox.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "flyte-sandbox.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Name of inline ConfigMap containing additional configuration or overrides for Flyte
*/}}
{{- define "flyte-sandbox.configuration.inlineConfigMap" -}}
{{- printf "%s-extra-config" .Release.Name -}}
{{- end }}

{{/*
Name of inline ConfigMap containing additional cluster resource templates
*/}}
{{- define "flyte-sandbox.clusterResourceTemplates.inlineConfigMap" -}}
{{- printf "%s-extra-cluster-resource-templates" .Release.Name -}}
{{- end }}

{{/*
Name of PersistentVolume and PersistentVolumeClaim for PostgreSQL database
*/}}
{{- define "flyte-sandbox.persistence.dbVolumeName" -}}
{{- printf "%s-db-storage" .Release.Name -}}
{{- end }}

{{/*
Name of PersistentVolume and PersistentVolumeClaim for Minio
*/}}
{{- define "flyte-sandbox.persistence.minioVolumeName" -}}
{{- printf "%s-minio-storage" .Release.Name -}}
{{- end }}

{{/*
Selector labels for Buildkit
*/}}
{{- define "flyte-sandbox.buildkitSelectorLabels" -}}
{{ include "flyte-sandbox.selectorLabels" . }}
app.kubernetes.io/component: buildkit
{{- end }}

{{/*
Selector labels for Envoy proxy
*/}}
{{- define "flyte-sandbox.proxySelectorLabels" -}}
{{ include "flyte-sandbox.selectorLabels" . }}
app.kubernetes.io/component: proxy
{{- end }}

{{/*
Name of Envoy proxy configmap
*/}}
{{- define "flyte-sandbox.proxyConfigMapName" -}}
{{- printf "%s-proxy-config" .Release.Name -}}
{{- end }}

{{/*
Name of development-mode Flyte headless service
*/}}
{{- define "flyte-sandbox.localHeadlessService" -}}
{{- printf "%s-local" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end }}
