[Match]
Name={{ .Name }}
Type=vlan

[Network]
Description=VLAN {{ .Name }}

{{ if .DNS -}}
DNS={{ .DNS }}
{{- end -}}

[Address]
{{ range .Addresses -}}
Address={{ . }}
{{ end -}}
{{ if .DHCP -}}
DHCP=yes
{{ end -}}

{{ if .Gateway }}
[Route]
Gateway={{ .Gateway }}
{{ end -}}
