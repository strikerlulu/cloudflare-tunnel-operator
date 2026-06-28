{{- define "cloudflare-tunnel-operator.fullname" -}}
{{- .Chart.Name -}}
{{- end -}}

{{- define "cloudflare-tunnel-operator.labels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}
