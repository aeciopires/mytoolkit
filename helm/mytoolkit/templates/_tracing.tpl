{{/*
OpenTelemetry tracing env vars for a Deployment's ConfigMap.
Usage: {{ include "mytoolkit.tracingEnv" (dict "root" . "service" "mytoolkit") }}
*/}}
{{- define "mytoolkit.tracingEnv" -}}
{{- $t := .root.Values.tracing -}}
{{- if $t.enabled }}
{{- if not $t.endpoint }}
{{- fail "tracing.enabled=true requires tracing.endpoint (an OTLP/HTTP endpoint, e.g. http://alloy.monitoring:4318)" }}
{{- end }}
OTEL_EXPORTER_OTLP_ENDPOINT: {{ $t.endpoint | quote }}
OTEL_SERVICE_NAME: {{ .service | quote }}
OTEL_TRACES_SAMPLER: {{ $t.sampler | quote }}
{{- with $t.samplerArg }}
OTEL_TRACES_SAMPLER_ARG: {{ . | quote }}
{{- end }}
{{- with $t.resourceAttributes }}
OTEL_RESOURCE_ATTRIBUTES: {{ . | quote }}
{{- end }}
{{- end }}
{{- end }}
