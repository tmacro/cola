[Match]
Name={{ .Name }}
Type=vlan

[Network]
Description=VLAN {{ .Name }}

{{ if .DNS -}}
DNS={{ .DNS }}
{{- end -}}

[Address]
{{ if .Address -}}
Address={{ .Address }}
{{ end -}}
{{ if .DHCP -}}
DHCP=yes
{{ end -}}

{{ if .Gateway }}
[Route]
Gateway={{ .Gateway }}
{{ end -}}
