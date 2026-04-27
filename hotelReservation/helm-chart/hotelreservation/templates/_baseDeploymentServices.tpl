{{- define "hotelreservation.templates.baseDeploymentServices" }}
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    {{- include "hotel-reservation.labels" . | nindent 4 }}
    service: {{ .Values.name }}
    app: {{ .Values.name }}
  name: {{ .Values.name }}
spec:
  replicas: {{ .Values.replicas | default .Values.global.replicas }}
  selector:
    matchLabels:
      {{- include "hotel-reservation.selectorLabels" . | nindent 6 }}
      service: {{ .Values.name }}
      app: {{ .Values.name }}
  template:
    metadata:
      labels:
        {{- include "hotel-reservation.labels" . | nindent 8 }}
        service: {{ .Values.name }}
        app: {{ .Values.name }}
      {{- if hasKey $.Values "annotations" }}
      annotations:
        {{ tpl $.Values.annotations . | nindent 8 | trim }}
      {{- else if hasKey $.Values.global "annotations" }}
      annotations:
        {{ tpl $.Values.global.annotations . | nindent 8 | trim }}
      {{- end }}
    spec:
      containers:
      {{- with .Values.container }}
      - name: "{{ .name }}"
        {{- $registry := .dockerRegistry | default $.Values.global.dockerRegistry }}
        {{- if eq $registry "" }}{{- fail "global.dockerRegistry (or per-subchart container.dockerRegistry) must be set; see values.yaml for details" }}{{- end }}
        image: {{ $registry }}/{{ .image }}:{{ .imageVersion | default $.Values.global.defaultImageVersion }}
        imagePullPolicy: {{ .imagePullPolicy | default $.Values.global.imagePullPolicy }}
        ports:
        {{- range $cport := .ports }}
        - containerPort: {{ $cport.containerPort -}}
        {{ end }}
        {{- if or (hasKey . "environments") (hasKey $.Values.global.services "environments") }}
        env:
        - name: NODE_IP
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          {{- $otelEndpoint := (($.Values.global.otel).endpoint) | default "" }}
          {{- if $otelEndpoint }}
          value: {{ $otelEndpoint | quote }}
          {{- else }}
          value: "http://$(NODE_IP):4318"
          {{- end }}
        {{- if hasKey . "environments" }}
        {{- range $variable, $value := .environments }}
        {{- if ne $variable "OTEL_EXPORTER_OTLP_ENDPOINT" }}
        - name: {{ $variable }}
          value: {{ $value | quote }}
        {{- end }}
        {{- end }}
        {{- else }}
        {{- range $variable, $value := $.Values.global.services.environments }}
        {{- if ne $variable "OTEL_EXPORTER_OTLP_ENDPOINT" }}
        - name: {{ $variable }}
          value: {{ $value | quote }}
        {{- end }}
        {{- end }}
        {{- end }}
        {{- end }}
        {{- if .command}}
        command:
        - {{ .command }}
        {{- end -}}
        {{- if .args}}
        args:
        {{- range $arg := .args}}
        - {{ $arg }}
        {{- end -}}
        {{- end }}
        {{- if .resources }}
        resources:
          {{ tpl .resources $ | nindent 10 | trim }}
        {{- else if hasKey $.Values.global "resources" }}
        resources:
          {{ tpl $.Values.global.resources $ | nindent 10 | trim }}
        {{- end }}
        {{- if $.Values.configMaps }}        
        volumeMounts: 
        {{- range $configMap := $.Values.configMaps }}
        - name: {{ $.Values.name }}-config
          mountPath: {{ $configMap.mountPath }}
          subPath: {{ $configMap.name }}
        {{- end }}
        {{- end }}
      {{- end -}}
      {{- if $.Values.configMaps }}
      volumes:
      - name: {{ $.Values.name }}-config
        configMap:
          name: {{ $.Values.name }}
      {{- end }}
      {{- if hasKey .Values "topologySpreadConstraints" }}
      topologySpreadConstraints:
        {{ tpl .Values.topologySpreadConstraints . | nindent 6 | trim }}
      {{- else if hasKey $.Values.global  "topologySpreadConstraints" }}
      topologySpreadConstraints:
        {{ tpl $.Values.global.topologySpreadConstraints . | nindent 6 | trim }}
      {{- end }}
      hostname: {{ .Values.name }}
      restartPolicy: {{ .Values.restartPolicy | default .Values.global.restartPolicy}}
      {{- if .Values.affinity }}
      affinity: {{- toYaml .Values.affinity | nindent 8 }}
      {{- else if hasKey $.Values.global "affinity" }}
      affinity: {{- toYaml .Values.global.affinity | nindent 8 }}
      {{- end }}
      {{- if .Values.tolerations }}
      tolerations: {{- toYaml .Values.tolerations | nindent 8 }}
      {{- else if hasKey $.Values.global "tolerations" }}
      tolerations: {{- toYaml .Values.global.tolerations | nindent 8 }}
      {{- end }}
      {{- if .Values.nodeSelector }}
      nodeSelector: {{- toYaml .Values.nodeSelector | nindent 8 }}
      {{- else if hasKey $.Values.global "nodeSelector" }}
      nodeSelector: {{- toYaml .Values.global.nodeSelector | nindent 8 }}
      {{- end }}
{{- end}}
