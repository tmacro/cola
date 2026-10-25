[Match]
Name={{ .Name }}
Type=vlan

[Network]
Description=VLAN {{ .Name }}
{{- if .DHCP }}
DHCP=yes
{{- end }}
{{- if .DNS }}
DNS={{ .DNS }}
{{- end }}
{{- range .Addresses }}

[Address]
Address={{ . }}
{{- end }}
{{- if .Gateway }}

[Route]
Gateway={{ .Gateway }}
{{- end }}
