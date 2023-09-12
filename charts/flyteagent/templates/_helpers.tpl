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


{{- define "flyteagent.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "flyteagent.selectorLabels" -}}
app.kubernetes.io/name: {{ template "flyteagent.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "flyteagent.labels" -}}
{{ include "flyteagent.selectorLabels" . }}
helm.sh/chart: {{ include "flyte.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

# Optional blocks for secret mount

{{- define "agentSecret.volume" -}}
- name: {{ include "flyte.name" . }}
  secret:
    secretName: {{ include "flyte.name" . }}
{{- end }}

{{- define "agentSecret.volumeMount" -}}
- mountPath: /etc/secrets
  name: {{ include "flyte.name" . }}
{{- end }}

{{- define "flyteagent.servicePort" -}}
{{ include .Values.ports.containerPort}}
{{- end }}