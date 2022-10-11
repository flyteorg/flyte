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
Common labels
*/}}
{{- define "flyte-binary.labels" -}}
helm.sh/chart: {{ include "flyte-binary.chart" . }}
{{ include "flyte-binary.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "flyte-binary.selectorLabels" -}}
app.kubernetes.io/name: {{ include "flyte-binary.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

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
Extra Minio connection settings
*/}}
{{- define "flyte-binary.minioExtraConnectionSettings" -}}
access-key: minio
auth-type: accesskey
secret-key: miniostorage
disable-ssl: true
endpoint: "http://localhost:30002"
{{- end }}

{{/*
Extra Minio env vars for propeller
*/}}
{{- define "flyte-binary.minioExtraEnvVars" -}}
- FLYTE_AWS_ENDPOINT: "http://minio.minio:9000"
- FLYTE_AWS_ACCESS_KEY_ID: minio
- FLYTE_AWS_SECRET_ACCESS_KEY: miniostorage
{{- end }}

{{/*
For creating a K8s secret for the database password
*/}}
{{- define "flyte-binary.database.secret" -}}
apiVersion: v1
kind: Secret
metadata:
  name: flyte-db-pass
type: Opaque
stringData:
  pass.txt: "{{ .Values.database.password }}"
{{- end }}

{{- define "flyte-binary.database.secretvol" -}}
- name: db-pass
  secret:
    secretName: flyte-db-pass
{{- end }}

{{- define "flyte-binary.database.secretvolmount" -}}
- mountPath: /etc/db
  name: db-pass
{{- end }}

{{- define "flyte-binary.database.secretpathconfig" -}}
passwordPath: /etc/db/pass.txt
{{- end }}



