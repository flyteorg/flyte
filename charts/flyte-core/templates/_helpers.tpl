{{/* vim: set filetype=mustache: */}}

{{- define "flyte.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "flyte.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "flyte.namespace" -}}
{{- default .Release.Namespace .Values.forceNamespace | trunc 63 | trimSuffix "-" -}}
{{- end -}}


{{- define "flyteadmin.name" -}}
flyteadmin
{{- end -}}

{{- define "flyteadmin.selectorLabels" -}}
app.kubernetes.io/name: {{ template "flyteadmin.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "flyteadmin.labels" -}}
{{ include "flyteadmin.selectorLabels" . }}
helm.sh/chart: {{ include "flyte.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "flyteadmin.podLabels" -}}
{{ include "flyteadmin.labels" . }}
{{- with .Values.flyteadmin.podLabels }}
{{ toYaml . }}
{{- end }}
{{- end -}}

{{- define "flytescheduler.name" -}}
flytescheduler
{{- end -}}

{{- define "flytescheduler.selectorLabels" -}}
app.kubernetes.io/name: {{ template "flytescheduler.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}


{{- define "flytescheduler.labels" -}}
{{ include "flytescheduler.selectorLabels" . }}
helm.sh/chart: {{ include "flyte.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "flytescheduler.podLabels" -}}
{{ include "flytescheduler.labels" . }}
{{- with .Values.flytescheduler.podLabels }}
{{ toYaml . }}
{{- end }}
{{- end -}}

{{- define "flyteclusterresourcesync.name" -}}
flyteclusterresourcesync
{{- end -}}

{{- define "flyteclusterresourcesync.selectorLabels" -}}
app.kubernetes.io/name: {{ template "flyteclusterresourcesync.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "flyteclusterresourcesync.labels" -}}
{{ include "flyteclusterresourcesync.selectorLabels" . }}
helm.sh/chart: {{ include "flyte.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "flyteclusterresourcesync.podLabels" -}}
{{ include "flyteclusterresourcesync.labels" . }}
{{- with .Values.cluster_resource_manager.podLabels }}
{{ toYaml . }}
{{- end }}
{{- end -}}
{{- define "datacatalog.name" -}}
datacatalog
{{- end -}}

{{- define "datacatalog.selectorLabels" -}}
app.kubernetes.io/name: {{ template "datacatalog.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "datacatalog.labels" -}}
{{ include "datacatalog.selectorLabels" . }}
helm.sh/chart: {{ include "flyte.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "datacatalog.podLabels" -}}
{{ include "datacatalog.labels" . }}
{{- with .Values.datacatalog.podLabels }}
{{ toYaml . }}
{{- end }}
{{- end -}}

{{- define "flytepropeller.name" -}}
flytepropeller
{{- end -}}

{{- define "flytepropeller.selectorLabels" -}}
app.kubernetes.io/name: {{ template "flytepropeller.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "flytepropeller.labels" -}}
{{ include "flytepropeller.selectorLabels" . }}
helm.sh/chart: {{ include "flyte.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "flytepropeller.podLabels" -}}
{{ include "flytepropeller.labels" . }}
{{- with .Values.flytepropeller.podLabels }}
{{ toYaml . }}
{{- end }}
{{- end -}}

{{- define "flytepropeller-manager.name" -}}
flytepropeller-manager
{{- end -}}

{{- define "flytepropeller-manager.selectorLabels" -}}
app.kubernetes.io/name: {{ template "flytepropeller-manager.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "flytepropeller-manager.labels" -}}
{{ include "flytepropeller-manager.selectorLabels" . }}
helm.sh/chart: {{ include "flyte.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "flytepropeller-manager.podLabels" -}}
{{ include "flytepropeller-manager.labels" . }}
{{- with .Values.flytepropeller.podLabels }}
{{ toYaml . }}
{{- end }}
{{- end -}}

{{- define "flyte-pod-webhook.name" -}}
flyte-pod-webhook
{{- end -}}


{{- define "flyteconsole.name" -}}
flyteconsole
{{- end -}}

{{- define "flyteconsole.selectorLabels" -}}
app.kubernetes.io/name: {{ template "flyteconsole.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "flyteconsole.labels" -}}
{{ include "flyteconsole.selectorLabels" . }}
helm.sh/chart: {{ include "flyte.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "flyteconsole.podLabels" -}}
{{ include "flyteconsole.labels" . }}
{{- with .Values.flyteconsole.podLabels }}
{{ toYaml . }}
{{- end }}
{{- end -}}

# Optional blocks for secret mount

{{- define "databaseSecret.volume" -}}
{{- with .Values.common.databaseSecret.name -}}
- name: {{ . }}
  secret:
    secretName: {{ . }}
{{- end }}
{{- end }}

{{- define "databaseSecret.volumeMount" -}}
{{- with .Values.common.databaseSecret.name -}}
- mountPath: /etc/db
  name: {{ . }}
{{- end }}
{{- end }}

{{- define "storage.base" -}}
storage:
{{- if eq .Values.storage.type "s3" }}
  type: s3
  container: {{ .Values.storage.bucketName | quote }}
  connection:
    auth-type: {{ .Values.storage.s3.authType }}
    region: {{ .Values.storage.s3.region }}
    {{- if eq .Values.storage.s3.authType "accesskey" }}
    access-key: {{ .Values.storage.s3.accessKey }}
    secret-key: {{ .Values.storage.s3.secretKey }}
    {{- end }}
{{- else if eq .Values.storage.type "gcs" }}
  type: stow
  stow:
    kind: google
    config:
      json: ""
      project_id: {{ .Values.storage.gcs.projectId }}
      scopes: https://www.googleapis.com/auth/cloud-platform
  container: {{ .Values.storage.bucketName | quote }}
{{- else if eq .Values.storage.type "sandbox" }}
  type: minio
  container: {{ .Values.storage.bucketName | quote }}
  stow:
    kind: s3
    config:
      access_key_id: minio
      auth_type: accesskey
      secret_key: miniostorage
      disable_ssl: true
      endpoint: http://minio.{{ .Release.Namespace }}.svc.cluster.local:9000
      region: us-east-1
  signedUrl:
    stowConfigOverride:
      endpoint: http://minio.{{ .Release.Namespace }}.svc.cluster.local:9000
{{- else if eq .Values.storage.type "custom" }}
{{- with .Values.storage.custom -}}
  {{ tpl (toYaml .) $ | nindent 2 }}
{{- end }}
{{- end }}
{{- end }}

{{- define "storage" -}}
{{ include "storage.base" .}}
  enable-multicontainer: {{ .Values.storage.enableMultiContainer }}
  limits:
    maxDownloadMBs: {{ .Values.storage.limits.maxDownloadMBs }}
{{- end }}
