# testinfra: local acceptance-test VM

This directory provisions a throwaway OpenWrt VM (via libvirt/KVM) to run this provider's acceptance test suite (`internal/acceptance/`) against, without needing a physical router.

From the repository root:

```bash
make testinfra-image  # build the OpenWrt VM image (once; cached under build/)
make testinfra-up     # provision the VM and wait for it to be ready
make testacc          # run the acceptance suite against the VM
make testinfra-down   # destroy the VM
```

## Prerequisites

- `qemu-kvm`, `libvirt-daemon-system`, `libvirt-clients` (or your distro's equivalents), plus a working `qemu:///system` connection - for the `dmacvicar/libvirt` Terraform provider used by `main.tf`.
- `podman` and `qemu-img` - used by `build-image.sh` to build the VM image via OpenWrt's ImageBuilder container.
- `terraform` and `curl` on your `PATH`.

## VM topology

`build-image.sh` builds the image from OpenWrt 24.10.2 (profile `generic`) with the packages needed for the provider's JSON-RPC client (`luci-mod-rpc`, `luci-lib-ipkg`, `luci-compat`, `uhttpd`, `luci-theme-bootstrap`) plus WireGuard and simulated-wireless support (`kmod-wireguard`, `wireguard-tools`, `kmod-mac80211-hwsim`, `wpad-mbedtls`, `wireless-tools`). See the comments above `PACKAGES=` in that script for why each package is needed.

On first boot (`image/files/etc/uci-defaults/99-acceptance-setup`):

- the root password is set to `root`
- `eth0` = `wan` (dhcp)
- `eth1` = `lan`/mgmt, static `192.168.56.2/24` - this is the address `OPENWRT_HOST` and the `openwrt_host` output (below) point at
- `eth2`-`eth5` are left unconfigured ("spare") for bridge/VLAN/network device acceptance tests to attach to
- `wifi config` populates `/etc/config/wireless` for the `radio0` (`mac80211_hwsim`) simulated radio

`main.tf` maps these 1:1 onto libvirt networks via MAC addresses `52:54:00:12:34:60`-`65` (`wan`, `mgmt`, `spare0`-`spare3`).

`internal/acceptance/testconfig.vm.yaml` describes this topology to the test suite - `make testacc` installs it as `internal/acceptance/testconfig.yaml` so `RequireTestConfig`-gated tests run against the VM by default.

## Variables

| Variable    | Default            | Description |
|-------------|--------------------|-------------|
| `memory_mb` | `256`              | VM memory in MiB |
| `vcpu`      | `1`                | VM vCPU count |
| `wan_cidr`  | `192.168.57.0/24`  | CIDR for the `wan` (`eth0`) libvirt network |
| `mgmt_cidr` | `192.168.56.0/24`  | CIDR for the `lan`/mgmt (`eth1`) libvirt network - must match the static address set by `build-image.sh`'s uci-defaults |

The defaults work out of the box; override with a `terraform.tfvars` file in this directory if needed.

## Outputs

- `openwrt_host` - base URL of the VM's JSON-RPC API, e.g. `http://192.168.56.2`. `make testacc` uses this as `OPENWRT_HOST`, with `OPENWRT_USERNAME=root` / `OPENWRT_PASSWORD=root`.

## Rebuilding the image

`build-image.sh --force` rebuilds `build/openwrt-acceptance.qcow2`, but `libvirt_volume.openwrt_base`'s `create.content.url` only uploads that file once. To pick up a rebuilt image on an already-provisioned VM:

```bash
cd testinfra
terraform apply \
  -replace=libvirt_volume.openwrt_base \
  -replace=libvirt_volume.openwrt_disk \
  -replace=libvirt_domain.openwrt
```

All three replacements are needed - replacing only the base volume leaves the running VM's COW disk, and thus its already-booted rootfs, untouched.

## Files

- `build-image.sh` - builds the qcow2 image via OpenWrt's ImageBuilder (podman).
- `wait-for-ready.sh` - polls the VM's JSON-RPC API until it responds (used by `make testinfra-up`).
- `providers.tf`, `main.tf`, `variables.tf`, `outputs.tf` - the `dmacvicar/libvirt` Terraform module.
- `image/files/` - overlay files baked into the image (uci-defaults, kernel module config).
