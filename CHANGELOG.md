# Changelog

All notable changes to the terraform-provider-openwrt will be documented in this file.

## 0.1.0 (Unreleased)

### Added

- Initial release
- Network device resource (openwrt_network_device)
- Network interface resource (openwrt_network_interface)
- Firewall zone resource (openwrt_firewall_zone)
- DHCP pool resource (openwrt_dhcp_pool)
- System LED resource (openwrt_system_led)
- Dropbear SSH resource (openwrt_dropbear)
- Wireless device resource (openwrt_wireless_device)
- Wireless interface resource (openwrt_wireless_iface)
- Sys hostname data source (openwrt_sys_hostname)
- Sys wireless data source (openwrt_sys_wireless)
- Sys wireless info data source (openwrt_sys_wireless_info)
- Sys reboot resource (openwrt_sys_reboot) - Triggers device reboot with optional delay
- Sys modprobe resource (openwrt_sys_modprobe) - Loads/unloads kernel modules with optional parameters
- Sys wireless radios data source (openwrt_sys_wireless_radios) - Discovers available wireless radio devices
- Sys wireless ifaces data source (openwrt_sys_wireless_ifaces) - Discovers wireless interfaces
- Supports reading configuration from environment variables
- ipkg_package resource: Added autoremove, force_remove, and update flags

### Fixed

- **wireless_interface**: Fixed Read function to use UCIForeach with `.name` matching instead of UCIGetAll
- **wireless_device**: Fixed Read function to use UCIForeach with `.name` matching instead of UCIGetAll
- **network_interface**: Fixed Create to use UCISetSection for named UCI sections; Fixed Read to search by `.name` (section name)
- **network_bridge_vlan**: Fixed Read to match by device+vlan; Fixed Create to use anonymous sections for OpenWrt 24.10 compatibility; Fixed Delete to find section by device+vlan
- **UCISetSection**: Fixed parameter format to use 3 separate params for the luci.model.uci.set() API

### Verified Working

- Network interface resources create properly named UCI sections (e.g., "guest", "iot") instead of anonymous sections (e.g., "cfg086d96")
- Network interface Read correctly finds resources by section name
- Network interface Destroy correctly removes named sections
- **network_bridge_vlan**: Creates, reads, and deletes bridge VLANs on OpenWrt using anonymous sections (e.g., vlan30)
- All unit tests pass

### Notes

- Live network interfaces (from `ip link`) may persist after UCI deletion due to OpenWrt network subsystem caching. Run `/etc/init.d/network reload` or reboot to fully remove stale interfaces.