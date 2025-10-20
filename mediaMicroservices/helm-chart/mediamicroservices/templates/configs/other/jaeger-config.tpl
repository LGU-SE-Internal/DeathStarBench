{{- define "mediamicroservices.templates.other.jaeger-config.yml"  }}
disabled: {{ .Values.global.otel.disabled }}
endpoint: "{{ .Values.global.otel.endpoint }}"
samplerParam: {{ .Values.global.otel.samplerParam }}
{{- end }}