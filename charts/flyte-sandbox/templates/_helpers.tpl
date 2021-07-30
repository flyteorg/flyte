{{/* vim: set filetype=mustache: */}}

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
