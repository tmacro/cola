system {
    hostname = "node"
}

extension "node-exporter" {
    version = "1.7.0-1"
    arch = "x86-64"
    bakery_url = "https://s3.binha.us/sysexts/images/"
}

extension "vmagent" {
    version = "1.101.0-1"
    arch = "x86-64"
    bakery_url = "https://s3.binha.us/sysexts/images/"
}

extension "consul" {
    version = "1.18.1-3"
    arch = "x86-64"
    bakery_url = "https://s3.binha.us/sysexts/images/"
}
