{{/*
Expand the name of the application.
*/}}
{{- define "k8s-deletion-inspector.name" -}}
{{- default .Chart.Name .Values.nameOverride -}}
{{- end -}}

{{/*
Expand the full name of the chart.
*/}}
{{- define "k8s-deletion-inspector.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "k8s-deletion-inspector.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "k8s-deletion-inspector.labels" -}}
helm.sh/chart: {{ include "k8s-deletion-inspector.chart" . }}
{{ include "k8s-deletion-inspector.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "k8s-deletion-inspector.selectorLabels" -}}
app.kubernetes.io/name: {{ include "k8s-deletion-inspector.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create a default name for the ServiceAccount
*/}}
{{- define "k8s-deletion-inspector.serviceAccountName" -}}
{{- if .Values.serviceAccount.name -}}
{{- .Values.serviceAccount.name -}}
{{- else -}}
{{- include "k8s-deletion-inspector.fullname" . -}}
{{- end -}}
{{- end -}}
