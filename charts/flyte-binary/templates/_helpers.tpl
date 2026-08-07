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
Console fully qualified name
*/}}
{{- define "flyte-binary.console.fullname" -}}
{{- printf "%s-console" (include "flyte-binary.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Console service name
*/}}
{{- define "flyte-binary.console.serviceName" -}}
{{- include "flyte-binary.console.fullname" . -}}
{{- end -}}

{{/*
Console selector labels
*/}}
{{- define "flyte-binary.console.selectorLabels" -}}
{{ include "flyte-binary.baseLabels" . }}
app.kubernetes.io/component: console
{{- end -}}

{{/*
Console common labels
*/}}
{{- define "flyte-binary.console.labels" -}}
helm.sh/chart: {{ include "flyte-binary.chart" . }}
{{ include "flyte-binary.console.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

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
Whether the configuration Secret carries a storage credentials file (013-storage-secrets.yaml)
for co-pilot to mount. Must stay in lockstep with the conditions guarding that key in
config-secret.yaml: naming a key that is never written leaves every task pod stuck in
ContainerCreating. secretKeyPath is excluded deliberately — it names a file that exists only
in this deployment's own container, so mounting it into a task pod would stop co-pilot from
starting; such deployments supply their own Secret via storage.copilotStorageSecretRef.
*/}}
{{- define "flyte-binary.configuration.copilotStorageSecretRendered" -}}
{{- if and (eq "s3" .Values.configuration.storage.provider) (eq "accesskey" .Values.configuration.storage.providerConfig.s3.authType) -}}
{{- .Values.configuration.storage.providerConfig.s3.secretKey -}}
{{- end -}}
{{- end -}}

{{/*
Whether the projected files would give co-pilot a *complete* storage configuration. Co-pilot
takes the stow config all-or-nothing — once it reads the mounted files, nothing is passed on
its command line — so an incomplete mount leaves it with no credentials at all rather than
falling back. Configuring it is therefore gated on this.

Complete when: the operator named the sources themselves; or the backend keeps everything in
003-storage.yaml (gcs, azure); or S3 needs no credentials (authType=iam, resolved ambiently);
or S3's credentials are in 013-storage-secrets.yaml.

Incomplete for S3 with secretKeyPath and no copilotStorageSecretRef: 003-storage.yaml carries
no credentials and the secret lives in a file only this deployment's container has. Those
deployments keep receiving the stow config on the co-pilot command line, as before.
*/}}
{{- define "flyte-binary.configuration.copilotStorageComplete" -}}
{{- with .Values.configuration.storage -}}
{{- if or .copilotStorageConfigMapRef .copilotStorageSecretRef -}}
true
{{- else if ne "s3" .provider -}}
true
{{- else if ne "accesskey" .providerConfig.s3.authType -}}
true
{{- else if .providerConfig.s3.secretKey -}}
true
{{- end -}}
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
Get the Flyte service HTTP port.
*/}}
{{- define "flyte-binary.service.http.port" -}}
{{- default 8090 .Values.service.ports.http -}}
{{- end -}}

{{/*
Get the Flyte API paths for ingress.
*/}}
{{- define "flyte-binary.ingress.apiPaths" -}}
- /flyteidl2.workflow.RunService
- /flyteidl2.workflow.RunService/*
- /flyteidl2.workflow.RunLogsService
- /flyteidl2.workflow.RunLogsService/*
- /flyteidl2.task.TaskService
- /flyteidl2.task.TaskService/*
- /flyteidl2.workflow.TranslatorService
- /flyteidl2.workflow.TranslatorService/*
- /flyteidl2.dataproxy.DataProxyService
- /flyteidl2.dataproxy.DataProxyService/*
- /flyteidl2.cluster.ClusterService
- /flyteidl2.cluster.ClusterService/*
- /flyteidl2.secret.SecretService
- /flyteidl2.secret.SecretService/*
- /flyteidl2.project.ProjectService
- /flyteidl2.project.ProjectService/*
- /flyteidl2.app.AppService
- /flyteidl2.app.AppService/*
- /flyteidl2.app.AppLogsService
- /flyteidl2.app.AppLogsService/*
- /flyteidl2.trigger.TriggerService
- /flyteidl2.trigger.TriggerService/*
- /flyteidl2.auth.IdentityService
- /flyteidl2.auth.IdentityService/*
- /flyteidl2.settings.SettingsService
- /flyteidl2.settings.SettingsService/*
{{- end -}}

{{/*
Get the Flyte auth-discovery paths for ingress. These are unauthenticated:
clients must reach them before they hold a token (OAuth server metadata and the
auth metadata service). IdentityService and SettingsService require auth and live
in apiPaths instead.
*/}}
{{- define "flyte-binary.ingress.wellknownPaths" -}}
- /.well-known/oauth-authorization-server
- /flyteidl2.auth.AuthMetadataService
- /flyteidl2.auth.AuthMetadataService/*
{{- end -}}

{{/*
Get the Flyte webhook service name.
*/}}
{{- define "flyte-binary.webhook.serviceName" -}}
{{- printf "%s-webhook" (include "flyte-binary.fullname" .) -}}
{{- end -}}

{{- define "flyte-binary.webhook.headlessServiceName" -}}
{{- printf "%s-webhook-headless" (include "flyte-binary.fullname" .) -}}
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
