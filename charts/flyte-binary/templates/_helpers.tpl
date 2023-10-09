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
Base labels
*/}}
{{- define "flyte-binary.baseLabels" -}}
app.kubernetes.io/name: {{ include "flyte-binary.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "flyte-binary.labels" -}}
helm.sh/chart: {{ include "flyte-binary.chart" . }}
{{ include "flyte-binary.baseLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "flyte-binary.selectorLabels" -}}
{{ include "flyte-binary.baseLabels" . }}
app.kubernetes.io/component: flyte-binary
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
Flag to use external configuration.
*/}}
{{- define "flyte-binary.configuration.externalConfiguration" -}}
{{- or .Values.configuration.externalConfigMap .Values.configuration.externalSecretRef -}}
{{- end -}}

{{/*
Get the Flyte configuration ConfigMap name.
*/}}
{{- define "flyte-binary.configuration.configMapName" -}}
{{- printf "%s-config" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{/*
Get the Flyte configuration Secret name.
*/}}
{{- define "flyte-binary.configuration.configSecretName" -}}
{{- printf "%s-config-secret" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{/*
Get the Flyte user data prefix.
*/}}
{{- define "flyte-binary.configuration.storage.userDataPrefix" -}}
{{- $userDataContainer := required "User data container required" .Values.configuration.storage.userDataContainer -}}
{{- if eq "s3" .Values.configuration.storage.provider -}}
{{- printf "s3://%s/data" $userDataContainer -}}
{{- else if eq "gcs" .Values.configuration.storage.provider -}}
{{- printf "gs://%s/data" $userDataContainer -}}
{{- end -}}
{{- end -}}

{{/*
Get the Flyte logging configuration.
*/}}
{{- define "flyte-binary.configuration.logging.plugins" -}}
{{- with .Values.configuration.logging.plugins -}}
kubernetes-enabled: {{ .kubernetes.enabled }}
{{- if .kubernetes.enabled }}
kubernetes-template-uri: {{ required "Template URI required for Kubernetes logging plugin" .kubernetes.templateUri }}
{{- end }}
cloudwatch-enabled: {{ .cloudwatch.enabled }}
{{- if .cloudwatch.enabled }}
cloudwatch-template-uri: {{ required "Template URI required for CloudWatch logging plugin" .cloudwatch.templateUri }}
{{- end }}
stackdriver-enabled: {{ .stackdriver.enabled }}
{{- if .stackdriver.enabled }}
stackdriver-template-uri: {{ required "Template URI required for stackdriver logging plugin" .stackdriver.templateUri }}
{{- end }}
{{- if .custom }}
templates: {{- toYaml .custom | nindent 2 -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Get the Secret name for Flyte admin authentication secrets.
*/}}
{{- define "flyte-binary.configuration.auth.adminAuthSecretName" -}}
{{- printf "%s-admin-auth" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{/*
Get the Secret name for Flyte authentication client secrets.
*/}}
{{- define "flyte-binary.configuration.auth.clientSecretName" -}}
{{- printf "%s-client-secrets" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{/*
Get the Flyte cluster resource templates ConfigMap name.
*/}}
{{- define "flyte-binary.clusterResourceTemplates.configMapName" -}}
{{- printf "%s-cluster-resource-templates" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{/*
Get the Flyte HTTP service name
*/}}
{{- define "flyte-binary.service.http.name" -}}
{{- printf "%s-http" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{/*
Get the Flyte GRPC service name
*/}}
{{- define "flyte-binary.service.grpc.name" -}}
{{- printf "%s-grpc" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{/*
Get the Flyte service HTTP port.
*/}}
{{- define "flyte-binary.service.http.port" -}}
{{- default 8088 .Values.service.ports.http -}}
{{- end -}}

{{/*
Get the Flyte service GRPC port.
*/}}
{{- define "flyte-binary.service.grpc.port" -}}
{{- default 8089 .Values.service.ports.grpc -}}
{{- end -}}

{{/*
Get the Flyte agent service GRPC port.
*/}}
{{- define "flyte-binary.flyteagent.grpc.port" -}}
{{- default 8000 .Values.service.ports.grpc -}}
{{- end -}}


{{/*
Get the Flyte webhook service name.
*/}}
{{- define "flyte-binary.webhook.serviceName" -}}
{{- printf "%s-webhook" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{/*
Get the Flyte webhook secret name.
*/}}
{{- define "flyte-binary.webhook.secretName" -}}
{{- printf "%s-webhook-secret" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{/*
Get the Flyte ClusterRole name.
*/}}
{{- define "flyte-binary.rbac.clusterRoleName" -}}
{{- printf "%s-cluster-role" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{/*
Get the name of the Flyte Agent Deployment.
*/}}
{{- define "flyte-binary.agent.name" -}}
{{- printf "%s-agent" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{/*
Flyte Agent selector labels
*/}}
{{- define "flyte-binary.agent.selectorLabels" -}}
{{ include "flyte-binary.baseLabels" . }}
app.kubernetes.io/component: agent
{{- end }}
