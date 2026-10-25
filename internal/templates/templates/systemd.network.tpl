[Match]
{{- if .Name }}
Name={{ .Name }}
{{- end }}
{{- if .MACAddress }}
MACAddress={{ .MACAddress }}
{{- end }}

[Network]
{{- if .DHCP }}
DHCP=yes
{{- end }}
{{- range .Addresses }}
Address={{ . }}
{{- end }}
{{- if .Gateway }}
Gateway={{ .Gateway }}
{{- end }}
{{- if .DNS }}
DNS={{ .DNS }}
{{- end }}
{{- range .VLANs }}
VLAN={{ .Name }}
{{- end }}
{{- if .VLANs }}
LinkLocalAddressing=no
LLDP=no
EmitLLDP=no
IPv6AcceptRA=no
IPv6SendRA=no
{{- end }}
