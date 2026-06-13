variable "memory_mb" {
  description = "Memory (in MiB) allocated to the OpenWrt acceptance VM."
  type        = number
  default     = 256
}

variable "vcpu" {
  description = "Number of vCPUs allocated to the OpenWrt acceptance VM."
  type        = number
  default     = 1
}

variable "wan_cidr" {
  description = "CIDR for the VM's wan (eth0) libvirt network."
  type        = string
  default     = "192.168.57.0/24"
}

variable "mgmt_cidr" {
  description = "CIDR for the VM's lan/mgmt (eth1) libvirt network. Must match the static address configured by testinfra/build-image.sh's uci-defaults."
  type        = string
  default     = "192.168.56.0/24"
}
