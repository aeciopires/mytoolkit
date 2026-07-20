{{/*
Expand the name of the chart.
*/}}
{{- define "mytoolkit.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "mytoolkit.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Chart name and version label.
*/}}
{{- define "mytoolkit.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels.
*/}}
{{- define "mytoolkit.labels" -}}
helm.sh/chart: {{ include "mytoolkit.chart" . }}
{{ include "mytoolkit.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels.
*/}}
{{- define "mytoolkit.selectorLabels" -}}
app.kubernetes.io/name: {{ include "mytoolkit.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Fully qualified name for the optional MCP Deployment/Service — distinct
from mytoolkit.fullname's "-mcp" suffix, mytoolkit.name in
mytoolkit.mcp.selectorLabels below (matching "app.kubernetes.io/name" to a
different value than the main workload) so its pod selector can never
overlap with the main Deployment's selector, which is immutable once
created and therefore must never be touched by this addition.
*/}}
{{- define "mytoolkit.mcp.fullname" -}}
{{- printf "%s-mcp" (include "mytoolkit.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Selector labels for the optional MCP Deployment/Service.
*/}}
{{- define "mytoolkit.mcp.selectorLabels" -}}
app.kubernetes.io/name: {{ include "mytoolkit.name" . }}-mcp
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Common labels for the optional MCP Deployment/Service.
*/}}
{{- define "mytoolkit.mcp.labels" -}}
helm.sh/chart: {{ include "mytoolkit.chart" . }}
{{ include "mytoolkit.mcp.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/component: mcp
{{- end -}}

{{/*
ServiceAccount name to use.
*/}}
{{- define "mytoolkit.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "mytoolkit.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{/*
ServiceAccount name for the optional MCP Deployment. Falls back to the
shared mytoolkit.serviceAccountName (preserving pre-existing behavior)
unless mcp.serviceAccount.create opts into a dedicated identity.
*/}}
{{- define "mytoolkit.mcp.serviceAccountName" -}}
{{- if .Values.mcp.serviceAccount.create -}}
{{- default (include "mytoolkit.mcp.fullname" .) .Values.mcp.serviceAccount.name -}}
{{- else -}}
{{- include "mytoolkit.serviceAccountName" . -}}
{{- end -}}
{{- end -}}

{{/*
Render a piece of yaml that defines manifests
Usage:
{{ include "mytoolkit.tools.render" ( dict "value" .Values.path.to.value "context" $ ) }}
*/}}
{{- define "mytoolkit.tools.render" -}}
{{- $value := typeIs "string" .value | ternary .value (.value | toYaml) }}
{{- if contains "{{" $value }}
  {{- tpl $value .context }}
{{- else }}
  {{- $value }}
{{- end }}
{{- end -}}
