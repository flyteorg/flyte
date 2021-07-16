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


{{- define "redis.name" -}}
redis
{{- end -}}

{{- define "redis.selectorLabels" -}}
app.kubernetes.io/name: {{ template "redis.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "redis.labels" -}}
{{ include "redis.selectorLabels" . }}
helm.sh/chart: {{ include "flyte.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}


{{- define "postgres.name" -}}
postgres
{{- end -}}

{{- define "postgres.selectorLabels" -}}
app.kubernetes.io/name: {{ template "postgres.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "postgres.labels" -}}
{{ include "postgres.selectorLabels" . }}
helm.sh/chart: {{ include "flyte.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}


{{- define "minio.name" -}}
minio
{{- end -}}

{{- define "minio.selectorLabels" -}}
app.kubernetes.io/name: {{ template "minio.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "minio.labels" -}}
{{ include "minio.selectorLabels" . }}
helm.sh/chart: {{ include "flyte.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "pytorch-operator.name" -}}
pytorch-operator
{{- end -}}

{{- define "pytorch-operator.namespace" -}}
pytorch-operator
{{- end -}}

{{- define "pytorch-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ template "pytorch-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "pytorch-operator.labels" -}}
{{ include "pytorch-operator.selectorLabels" . }}
helm.sh/chart: {{ include "flyte.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
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
    auth-type: iam
    region: {{ .Values.storage.s3.region }}
{{- else if eq .Values.storage.type "gcs" }}
  type: stow
  stow:
    kind: google
    config:
      json: ""
      project_id: {{ .Values.storage.gcs.projectId }}
      scopes: https://www.googleapis.com/auth/devstorage.read_write
  container: {{ .Values.storage.bucketName | quote }}
{{- else if eq .Values.storage.type "sandbox" }}
  type: minio
  container: {{ .Values.storage.bucketName | quote }}
  connection:
    access-key: minio
    auth-type: accesskey
    secret-key: miniostorage
    disable-ssl: true
    endpoint: http://minio.{{ .Release.Namespace }}.svc.cluster.local:9000
    region: us-east-1
{{- else if eq .Values.storage.type "custom" }}
{{- with .Values.storage.custom -}}
  {{ toYaml . | nindent 2 }}
{{- end }}
{{- end }}
{{- end }}

{{- define "storage" -}}
{{ include "storage.base" .}}
  limits:
    maxDownloadMBs: 10
{{- end }}

{{- define "copilot.config" -}}
kind: ConfigMap
apiVersion: v1
metadata:
  name: flyte-data-config
  namespace: {{`{{ namespace }}`}}
data:
  config.yaml: | {{ include "storage.base" . | nindent 4 }}
      enable-multicontainer: true
{{- end }}
