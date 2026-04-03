{{/*
Expand the name of the chart.
*/}}
{{- define "hologram-mqtt.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "hologram-mqtt.fullname" -}}
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
{{- define "hologram-mqtt.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "hologram-mqtt.labels" -}}
helm.sh/chart: {{ include "hologram-mqtt.chart" . }}
{{ include "hologram-mqtt.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "hologram-mqtt.selectorLabels" -}}
app.kubernetes.io/name: {{ include "hologram-mqtt.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "hologram-mqtt.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "hologram-mqtt.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Hologram API key secret name
*/}}
{{- define "hologram-mqtt.hologramSecretName" -}}
{{- if .Values.hologram.existingSecret }}
{{- .Values.hologram.existingSecret }}
{{- else }}
{{- include "hologram-mqtt.fullname" . }}-hologram
{{- end }}
{{- end }}

{{/*
MQTT credentials secret name
*/}}
{{- define "hologram-mqtt.mqttSecretName" -}}
{{- if .Values.mqtt.existingSecret }}
{{- .Values.mqtt.existingSecret }}
{{- else }}
{{- include "hologram-mqtt.fullname" . }}-mqtt
{{- end }}
{{- end }}
