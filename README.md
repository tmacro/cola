# COLA - Container Linux Appliance

COLA is set of tools to build customized Flatcar Container Linux configurations and images.
It includes a transpiler to convert a high-level HCL based configuration to a low-level Ignition configuration.
It is still in early development and breaking changes are to be expected.

## Configuration Example

```hcl
system {
    hostname = "dhcp-server"
}

user "kea" {
  uid            = 100
  shell          = "/bin/false"
  no_create_home = true
}

container "kea" {
  image   = "docker.cloudsmith.io/isc/docker/kea-dhcp4:latest"
  restart = "always"
  cap_add = ["NET_ADMIN", "NET_RAW"]

  volume "/etc/kea" {
    source = "/etc/kea"
  }

  volume "/var/lib/kea" {
    source = "/var/lib/kea"
  }
}

directory "/var/lib/kea" {
  mode  = "0755"
  owner = "kea"
  group = "kea"
}

directory "/etc/kea" {
  mode  = "0755"
  owner = "kea"
  group = "kea"
}

file "/etc/kea/kea-dhcp4.conf" {
  mode   = "0644"
  owner  = "kea"
  group  = "kea"
  inline = <<-EOF
        {
            "Dhcp4": {
                "valid-lifetime": 4000,
                "renew-timer": 1000,
                "rebind-timer": 2000,
                "interfaces-config": {
                    "interfaces": ["eth0"],
                    "dhcp-socket-type": "raw"
                },
                "lease-database": {
                    "type": "memfile",
                    "persist": true,
                    "name": "/var/lib/kea/dhcp4.leases"
                },
                "subnet4": [
                    {
                        "id": 1,
                        "subnet": "10.29.0.1/24",
                        "pools": [{ "pool": "10.29.0.100-10.29.0.200" }]
                    }
                ],
                "option-data": [
                    {
                        "name": "routers",
                        "data": "10.29.0.1",
                    },
                    {
                        "name": "domain-name-servers",
                        "code": 6,
                        "space": "dhcp4",
                        "csv-format": true,
                        "data": "10.29.0.1"
                    }
                ],
                "loggers": [
                    {
                        "name": "*",
                        "severity": "DEBUG"
                    }
                ]
            }
        }
    EOF
}
```
