# Contributing

## Development Setup

```bash
# Clone the repository
git clone https://github.com/bucksbunni/terraform-provider-openwrt.git
cd terraform-provider-openwrt

# Install dependencies
go mod download

# Install tools for documentation
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
go install github.com/YakDriver/tfproviderdocs@latest
```

## Building and Testing

```bash
# Build the provider
make build

# Run tests
make test

# Run linter
make lint
```

## Acceptance Tests

The acceptance tests in `internal/acceptance/` exercise real resources against a live OpenWrt instance and are skipped unless `TF_ACC=1` is set.

For local development, `testinfra/` provisions a throwaway OpenWrt VM via libvirt/KVM:

```bash
make testinfra-image  # build the VM image (once)
make testinfra-up     # provision the VM
make testacc          # run the acceptance suite against it
make testinfra-down   # destroy the VM
```

See [testinfra/README.md](testinfra/README.md) for prerequisites and VM details, and run `go doc ./internal/acceptance` for environment variables and test-writing conventions.

## Documentation

### Generating Documentation

When adding new resources/data sources or changing schemas, regenerate documentation:

```bash
make docs
```

This runs `tfplugindocs` to auto-generate docs from the provider schema.

### Preserving Custom Content

Some documentation files contain custom content (examples, guides, additional context). After running `make docs`, you must manually restore custom content in:

- `docs/index.md`
- `docs/guides/wireless-setup.md`
- `docs/resources/wireless_device.md`
- `docs/resources/wireless_iface.md`
- `docs/resources/network_interface.md`

### Validating Documentation

Before submitting a PR, verify docs meet Terraform Registry standards:

```bash
make docs-check
```

This builds the provider and runs `tfproviderdocs check`.

## GitHub Actions

The documentation workflow (`.github/workflows/docs.yml`) runs on:

- Every PR modifying Go files or `.tfplugindocs.yml`
- Every push to `main`

It validates that documentation is up-to-date. If docs drift is detected, the workflow fails with a warning.

## Release Process

Releases are automated via GoReleaser. See `.goreleaser.yml` for configuration.

## Commit Messages

This project follows [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/); messages are validated in CI and drive the auto-generated release notes.

## Code Style

- Follow standard Go conventions
- Use meaningful variable names
- Add inline comments only when necessary
- Ensure `make lint` passes before submitting PRs
