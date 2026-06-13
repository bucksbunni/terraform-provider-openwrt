output "openwrt_host" {
  description = "Base URL of the OpenWrt acceptance VM's JSON-RPC API (lan/mgmt interface)."
  value       = "http://${cidrhost(var.mgmt_cidr, 2)}"
}
