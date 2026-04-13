{{/*
Expand the name of the chart.
*/}}
{{- define "flyte-demo.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "flyte-demo.fullname" -}}
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
{{- define "flyte-demo.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "flyte-demo.labels" -}}
helm.sh/chart: {{ include "flyte-demo.chart" . }}
{{ include "flyte-demo.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "flyte-demo.selectorLabels" -}}
app.kubernetes.io/name: {{ include "flyte-demo.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "flyte-demo.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "flyte-demo.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Name of inline ConfigMap containing additional configuration or overrides for Flyte
*/}}
{{- define "flyte-demo.configuration.inlineConfigMap" -}}
{{- printf "%s-extra-config" .Release.Name -}}
{{- end }}

{{/*
Name of inline ConfigMap containing additional cluster resource templates
*/}}
{{- define "flyte-demo.clusterResourceTemplates.inlineConfigMap" -}}
{{- printf "%s-extra-cluster-resource-templates" .Release.Name -}}
{{- end }}

{{/*
Name of PersistentVolume and PersistentVolumeClaim for PostgreSQL database
*/}}
{{- define "flyte-demo.persistence.dbVolumeName" -}}
{{- printf "%s-db-storage" .Release.Name -}}
{{- end }}

{{/*
Name of PersistentVolume and PersistentVolumeClaim for RustFS
*/}}
{{- define "flyte-demo.persistence.rustfsVolumeName" -}}
{{- printf "%s-rustfs-storage" .Release.Name -}}
{{- end }}


{{/*
Selector labels for console
*/}}
{{- define "flyte-demo.consoleSelectorLabels" -}}
{{ include "flyte-demo.selectorLabels" . }}
app.kubernetes.io/component: console
{{- end }}

{{/*
Name of development-mode Flyte headless service
*/}}
{{- define "flyte-demo.localHeadlessService" -}}
{{- printf "%s-local" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end }}
