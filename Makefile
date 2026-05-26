.PHONY: build test lint docs docs-check clean help

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

# Show help
help:
	@echo "Available targets:"
	@echo "  build       - Build the provider"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linter"
	@echo "  docs        - Generate documentation"
	@echo "  docs-check  - Validate documentation meets registry standards"
	@echo "  clean       - Clean build artifacts"
	@echo "  help        - Show this help message"