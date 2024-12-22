# COLA - Container Linux Appliance

COLA is set of tools to build customized Flatcar Container Linux configurations and images.
It includes a transpiler to convert a high-level HCL based configuration to a low-level Ignition configuration.
It is still in early development and breaking changes are to be expected.

## Block Types

### System

The `system` block is used to configure OS level settings.

```hcl
system {
    hostname = "dhcp-server"
    enable_tty_auto_login = true
}
```

#### Fields

- `hostname` (string): The hostname of the machine.
- `enable_tty_auto_login` (bool): Enable auto login on tty1.

### User

```hcl
user "janitor" {
    uid = 1000
    groups   = ["sudo", "docker"]
    home_dir = "/home/janitor"
    shell    = "/bin/bash"
    ssh_authorized_keys = [
        "ssh-ed25519 ...",
    ]
}
```

#### Fields

- `uid` (int): The user id.
- `groups` (list of strings): The groups the user belongs to.
- `home_dir` (string): The home directory of the user.
- `no_create_home` (bool): Do not create the home directory of the user.
- `shell` (string): The shell of the user.
- `ssh_authorized_keys` (list of strings): The SSH authorized keys of the user.

###  Service

```hcl
service "etcd-member.service" {
  enabled = true

  drop_in "20-clustering.conf" {
    inline  = <<-EOF
        [Unit]
        Requires=coreos-metadata.service
        After=coreos-metadata.service
        [Service]
        ExecStart=
        ExecStart=/usr/lib/coreos/etcd-wrapper $ETCD_OPTS \
            --listen-peer-urls="http://$${COREOS_CUSTOM_PRIVATE_IPV4}:2380" \
            --listen-client-urls="http://0.0.0.0:2379,http://0.0.0.0:4001" \
            --initial-advertise-peer-urls="http://$${COREOS_CUSTOM_PRIVATE_IPV4}:2380" \
            --advertise-client-urls="http://$${COREOS_CUSTOM_PRIVATE_IPV4}:2379,http://$${COREOS_CUSTOM_PRIVATE_IPV4}:4001" \
            --initial-cluster-token "$${INITIAL_CLUSTER_TOKEN}" \
            --initial-cluster "$${INITIAL_CLUSTER}" \
            --initial-cluster-state new
        EOF
    }
}
```

#### Fields

- `enabled` (bool): Enable the service.
- `drop_in` (list of drop_in): Drop-in configuration for the service.
- `source_path` (string): Path to the file containing the configuration. Relative to the configuration file.

#### Drop-in Fields

- `inline` (string): The content of the drop-in configuration.
- `source_path` (string): Path to the file containing the configuration. Relative to the configuration file.

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
