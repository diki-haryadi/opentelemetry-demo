{{- define "rctiplus.fullname" -}}
{{- .Chart.Name }}-{{ .Release.Name }}
{{- end -}}
