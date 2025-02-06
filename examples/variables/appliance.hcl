system {
  hostname = var.hostname
}

variable "hostname" {
  type = string
}

variable "timezone" {
  type = string
  default = "UTC"
}

variable "num_hosts" {
  type = number
}

file "/etc/testme.conf" {
  mode = "0644"
  inline = <<-EOF
    hostname = "${var.hostname}"
    timezone = "${var.timezone}"
    num = ${var.num_hosts}
  EOF
}
