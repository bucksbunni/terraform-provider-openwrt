.PHONY: build test lint docs docs-check clean help testinfra-image testinfra-up testinfra-down testacc

# Default target
default: build

# Build the provider
build:
	@echo "Building provider..."
	go build -o terraform-provider-openwrt .

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run linter (golangci-lint must be installed)
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Generate documentation
docs:
	@echo "Generating documentation..."
	./scripts/gen-docs.sh

# Check documentation validity (for CI)
docs-check: build
	@echo "Checking documentation..."
	tfproviderdocs check \
		--provider-name openwrt \
		--provider-source registry.terraform.io/bucksbunni/openwrt

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f terraform-provider-openwrt
	go clean

# Build the OpenWrt VM image for acceptance tests (skips if already built;
# use ./testinfra/build-image.sh --force to rebuild)
testinfra-image:
	@echo "Building OpenWrt acceptance VM image..."
	./testinfra/build-image.sh

# Provision the local libvirt/KVM acceptance VM and wait for it to be ready
testinfra-up:
	@echo "Bringing up testinfra VM..."
	cd testinfra && terraform init -input=false && terraform apply -auto-approve
	@echo "Waiting for testinfra VM to become ready..."
	./testinfra/wait-for-ready.sh "$$(cd testinfra && terraform output -raw openwrt_host)"

# Destroy the local libvirt/KVM acceptance VM
testinfra-down:
	@echo "Tearing down testinfra VM..."
	cd testinfra && terraform destroy -auto-approve

# Run acceptance tests against the testinfra VM (requires `make testinfra-up`)
testacc:
	@echo "Installing VM-matching testconfig.yaml..."
	cp internal/acceptance/testconfig.vm.yaml internal/acceptance/testconfig.yaml
	@echo "Running acceptance tests against testinfra VM..."
	TF_ACC=1 \
	OPENWRT_HOST="$$(cd testinfra && terraform output -raw openwrt_host)" \
	OPENWRT_USERNAME=root \
	OPENWRT_PASSWORD=root \
	go test ./internal/acceptance/... -v -count=1

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build the provider"
	@echo "  test            - Run tests"
	@echo "  lint            - Run linter"
	@echo "  docs            - Generate documentation"
	@echo "  docs-check      - Validate documentation meets registry standards"
	@echo "  clean           - Clean build artifacts"
	@echo "  testinfra-image - Build the OpenWrt acceptance VM image"
	@echo "  testinfra-up    - Provision the local acceptance VM"
	@echo "  testinfra-down  - Destroy the local acceptance VM"
	@echo "  testacc         - Run acceptance tests against the VM"
	@echo "  help            - Show this help message"