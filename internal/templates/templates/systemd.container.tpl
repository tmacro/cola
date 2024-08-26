[Unit]
Description=Container {{.Name}}
After=local-fs.target
After=network-online.target

[Container]
Image={{.Image}}
{{if .Args}}
Exec={{.Args | join " "}}
{{end}}
Network=host
{{range .Volumes}}
Volume={{.Source}}:{{.Target}}
{{end}}

[Install]
WantedBy=multi-user.target default.target
