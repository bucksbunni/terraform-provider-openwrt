terraform {
  required_providers {
    openwrt = {
      source  = "bucksbunni/openwrt"
      version = "~> 0.1.0"
    }
  }
}

provider "openwrt" {
  host     = var.openwrt_host
  username = var.openwrt_username
  password = var.openwrt_password
  insecure = var.insecure
}

variable "openwrt_host" {
  description = "OpenWrt LuCI URL"
  type        = string
  default     = "http://192.168.1.1"
}

variable "openwrt_username" {
  description = "OpenWrt username"
  type        = string
  default     = "root"
}

variable "openwrt_password" {
  description = "OpenWrt password"
  type        = string
  sensitive   = true
}

variable "insecure" {
  description = "Skip TLS certificate verification"
  type        = bool
  default     = true
}
