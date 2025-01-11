[Match]
{{ if .Name -}}
Name={{ .Name }}
{{- end }}

{{- if .MACAddress -}}
MACAddress={{ .MACAddress }}
{{- end }}

[Network]
{{- if .Gateway }}
Gateway={{ .Gateway }}
{{- end }}

{{- if .DNS -}}
DNS={{ .DNS }}
{{- end }}


{{- range .Addresses }}
Address={{ . }}
{{- end }}

{{ range .VLANs -}}
VLAN={{ .Name }}
{{ end -}}

{{- if len .VLANs -}}
LinkLocalAddressing=no
LLDP=no
EmitLLDP=no
IPv6AcceptRA=no
IPv6SendRA=no
{{ end -}}
