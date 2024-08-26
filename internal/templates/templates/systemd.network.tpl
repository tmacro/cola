[Match]
{{ if .Name }}Name={{ .Name }}{{ end }}

[Network]
{{ if .Gateway -}}
Gateway={{ .Gateway }}
{{ end -}}

{{- if .DNS -}}
DNS={{ .DNS }}
{{ end }}

{{- if .Address }}
Address={{ .Address }}
{{ end -}}

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
