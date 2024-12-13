{{ range . }}
{{ .Mode }} {{ .Target }} - - - - {{ .Source }}
{{ end }}
