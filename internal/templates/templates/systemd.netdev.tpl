[NetDev]
Name={{ .Name }}
Kind={{ .Kind }}

{{ if .Kind | eq "vlan" -}}
[VLAN]
Id={{ .ID }}
{{ end }}
