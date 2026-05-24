// Package acceptance provides acceptance tests for the OpenWrt Terraform provider.
// These tests require a real OpenWrt instance to run against.
//
// # Prerequisites
//
// Set the following environment variables before running:
//
//	OPENWRT_HOST      Base URL of the OpenWrt LuCI instance (e.g., http://192.168.1.1)
//	OPENWRT_USERNAME  OpenWrt username (typically 'root')
//	OPENWRT_PASSWORD  OpenWrt password
//
// Optional environment variables:
//
//	OPENWRT_INSECURE          Set to "true" to skip TLS verification
//	OPENWRT_SKIP_PRECHECK      Set to "true" to skip environment variable validation
//	OPENWRT_SKIP_CONNECTIVITY  Set to "true" to skip connectivity check
//
// # Setting Up a Test VM
//
// The testinfra/ directory contains Terraform configuration for provisioning
// an OpenWrt VM using libvirt/KVM. To set up a test VM:
//
//	# Install prerequisites (Debian/Ubuntu)
//	sudo apt install -y qemu-kvm libvirt-daemon-system libvirt-clients bridge-utils
//
//	# Initialize Terraform
//	cd testinfra
//	cp terraform.tfvars.example terraform.tfvars
//	# Edit terraform.tfvars with your settings
//	terraform init
//
//	# Provision the VM
//	terraform apply
//
//	# Get the VM IP address
//	OPENWRT_HOST="http://$(terraform output -raw openwrt_ip)"
//
//	# After tests, cleanup
//	terraform destroy
//
// For more details, see testinfra/README.md.
//
// # Running Tests
//
// Run all acceptance tests:
//
//	export OPENWRT_HOST="http://192.168.122.100"
//	export OPENWRT_USERNAME="root"
//	export OPENWRT_PASSWORD="your-password"
//	TF_ACC=1 go test ./internal/acceptance/... -v
//
// Run specific test:
//
//	TF_ACC=1 go test ./internal/acceptance/... -v -run TestAccFirewallZone
//
// Run tests with verbose output:
//
//	TF_ACC=1 go test ./internal/acceptance/... -v -count=1
//
// # Test Categories
//
// This package includes acceptance tests for:
//
//   - Firewall: zones, rules, forwarding, defaults
//   - Network: interfaces, devices, bridge vlans, globals, WireGuard
//   - DHCP: pools, DNSmasq, odhcpd, hosts
//   - System: system, NTP, LEDs, Dropbear
//   - Wireless: devices, interfaces
//   - uHTTPd: web server, certificates
//   - RPC: RPC daemon
//
// # Safety Considerations
//
// Acceptance tests are designed to be safe and non-destructive:
//
//   - Tests use unique prefixes (tf-acc-*) to identify test resources
//   - Tests avoid modifying existing configurations (e.g., default LAN, DHCP pools)
//   - Tests clean up created resources during teardown
//   - Use OPENWRT_TEST_NON_DEFAULT to explicitly test resources that modify existing configs
//
// # Test Patterns
//
// This package follows patterns from hashicorp/terraform-provider-aws and
// paultyng/terraform-provider-unifi:
//
//   - PreCheck functions for environment validation
//   - Provider factories for Terraform Plugin Framework
//   - Config helper functions for generating Terraform HCL
//   - Check functions for state verification
//   - Disappears tests for external deletion handling
//
// # CI/CD Integration
//
// For CI/CD pipelines, use environment variables:
//
//	script:
//	  - export OPENWRT_HOST="http://${{ vars.OPENWRT_IP }}"
//	  - export OPENWRT_USERNAME="root"
//	  - export OPENWRT_PASSWORD=${{ secrets.OPENWRT_PASSWORD }}
//	  - TF_ACC=1 go test ./internal/acceptance/... -v -count=1
package acceptance
