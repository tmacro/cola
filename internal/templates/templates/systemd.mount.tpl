[Unit]
Description=Mount {{.Where}}
After=network.target

[Mount]
What={{ .What }}
Where={{ .Where }}
Type={{ .Type }}
Options={{ .Options }}

[Install]
WantedBy=multi-user.target
