resource "libvirt_volume" "openwrt_base" {
  name = "openwrt-acceptance-base.qcow2"
  pool = "default"

  target = {
    format = { type = "qcow2" }
  }

  create = {
    content = {
      url = "${path.module}/build/openwrt-acceptance.qcow2"
    }
  }
}

resource "libvirt_volume" "openwrt_disk" {
  name     = "openwrt-acceptance-disk.qcow2"
  pool     = "default"
  capacity = libvirt_volume.openwrt_base.capacity

  target = {
    format = { type = "qcow2" }
  }

  backing_store = {
    path   = libvirt_volume.openwrt_base.path
    format = { type = "qcow2" }
  }
}

resource "libvirt_network" "wan" {
  name = "openwrt-acceptance-wan"

  forward = {
    mode = "nat"
  }

  ips = [{
    address = cidrhost(var.wan_cidr, 1)
    prefix  = tonumber(split("/", var.wan_cidr)[1])
    dhcp = {
      ranges = [{
        start = cidrhost(var.wan_cidr, 10)
        end   = cidrhost(var.wan_cidr, 199)
      }]
    }
  }]
}

resource "libvirt_network" "mgmt" {
  name = "openwrt-acceptance-mgmt"

  forward = {
    mode = "nat"
  }

  ips = [{
    address = cidrhost(var.mgmt_cidr, 1)
    prefix  = tonumber(split("/", var.mgmt_cidr)[1])
    # no `dhcp` block: eth1/lan gets its static 192.168.56.2/24 from
    # uci-defaults, not from this network's DHCP server.
  }]
}

resource "libvirt_network" "spare" {
  count = 4
  name  = "openwrt-acceptance-spare${count.index}"

  forward = {
    mode = "none"
  }
}

resource "libvirt_domain" "openwrt" {
  name        = "openwrt-acceptance"
  type        = "kvm"
  memory      = var.memory_mb
  memory_unit = "MiB"
  vcpu        = var.vcpu
  running     = true

  os = {
    type         = "hvm"
    type_arch    = "x86_64"
    boot_devices = [{ dev = "hd" }]
  }

  devices = {
    disks = [{
      device = "disk"
      driver = { name = "qemu", type = "qcow2" }
      source = {
        volume = {
          pool   = "default"
          volume = libvirt_volume.openwrt_disk.name
        }
      }
      target = { dev = "vda", bus = "virtio" }
    }]

    interfaces = concat(
      [
        {
          mac    = { address = "52:54:00:12:34:60" }
          model  = { type = "virtio" }
          source = { network = { network = libvirt_network.wan.name } }
        },
        {
          mac    = { address = "52:54:00:12:34:61" }
          model  = { type = "virtio" }
          source = { network = { network = libvirt_network.mgmt.name } }
        },
      ],
      [
        for i, net in libvirt_network.spare : {
          mac    = { address = format("52:54:00:12:34:6%d", 2 + i) }
          model  = { type = "virtio" }
          source = { network = { network = net.name } }
        }
      ]
    )

    consoles = [{
      target = { type = "serial", port = 0 }
    }]
  }
}
