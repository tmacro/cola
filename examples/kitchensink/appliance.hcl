system {
    hostname = "flatcartest"
}

user "janitor" {
    groups = ["sudo", "docker"]
    home_dir = "/home/janitor"
    shell = "/bin/bash"
    ssh_authorized_keys = [
        "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIChZagJMY4JZfplh0UC3weTuB8k6cJ6uWWqY6Ww18mR+ janitor@lab"
    ]
}

extension "node-exporter" {
    bakery_url = "https://s3.binha.us/sysexts/images/"
    version = "1.7.0-1"
    arch = "x86-64"
}

extension "vmagent" {
    bakery_url = "https://s3.binha.us/sysexts/images/"
    version = "1.101.0-1"
    arch = "x86-64"
}

extension "consul" {
    bakery_url = "https://s3.binha.us/sysexts/images/"
    version = "1.18.1-3"
    arch = "x86-64"
}

extension "coredns" {
    bakery_url = "https://s3.binha.us/sysexts/images/"
    version = "1.11.3-1"
    arch = "x86-64"
}

file "/etc/testme.conf" {
    mode = "0644"
    source_path = "test.conf"
}

directory "/etc/testme" {
    mode = "0755"
}
