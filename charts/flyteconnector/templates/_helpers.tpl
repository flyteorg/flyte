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


{{- define "flyteconnector.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "flyteconnector.selectorLabels" -}}
app.kubernetes.io/name: {{ template "flyteconnector.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "flyteconnector.labels" -}}
{{ include "flyteconnector.selectorLabels" . }}
helm.sh/chart: {{ include "flyte.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "flyteconnector.podLabels" -}}
{{ include "flyteconnector.labels" . }}
{{- with .Values.podLabels }}
{{ toYaml . }}
{{- end }}
{{- end -}}

# Optional blocks for secret mount

{{- define "connectorSecret.volume" -}}
- name: {{ include "flyte.name" . }}
  secret:
    secretName: {{ include "flyte.name" . }}
{{- end }}

{{- define "connectorSecret.volumeMount" -}}
- mountPath: /etc/secrets
  name: {{ include "flyte.name" . }}
{{- end }}

{{- define "flyteconnector.servicePort" -}}
{{ include .Values.ports.containerPort}}
{{- end }}
