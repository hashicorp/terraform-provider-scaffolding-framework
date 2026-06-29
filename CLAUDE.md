# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build       # Build the provider
make install     # Build and install to GOPATH
make test        # Run unit tests (with coverage, 120s timeout, 10 parallel)
make testacc     # Run acceptance tests (requires TF_ACC=1, hits real API)
make lint        # Run golangci-lint via tools/go.mod
make fmt         # Format with gofumpt
make generate    # Regenerate docs (runs terraform fmt on examples/, then tfplugindocs)
make update      # Pull git submodules
```

Run a single test:
```bash
go test -v -run TestAccBlockStorage ./internal/acctest/
go test -v -run TestFoo ./internal/provider/
```

Acceptance tests require a live SECA cluster. The `internal/acctest/provider_test.go` hardcodes endpoint URLs for a local cluster at `172.18.0.2:30081`.

## Architecture

This is a [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) provider for the EU Sovereign Cloud (SECA) platform, registered at `registry.terraform.io/eu-sovereign-cloud/seca`.

### Key layers

- **`main.go`** — entrypoint; supports `--debug` flag for Delve
- **`internal/provider/`** — all provider logic (single package)
- **`internal/acctest/`** — acceptance tests (separate package; uses `testAccProtoV6ProviderFactories`)
- **`tools/`** — separate Go module for dev tooling (golangci-lint, tfplugindocs, gofumpt)
- **`spec/`** — git submodule pointing to `github.com/eu-sovereign-cloud/spec`

### Client initialization (`clients.go`)

The provider initializes two SDK clients from `github.com/eu-sovereign-cloud/go-sdk/secapi`:
- `GlobalClient` — talks to global endpoints (region registry, authorization)
- `RegionalClient` — derived from `GlobalClient.NewRegionalClient()`; used by all resources/data sources

Both are passed to resources and data sources via `resp.DataSourceData` / `resp.ResourceData` as the `clients` struct.

### Resource/data source pattern

Each resource and data source follows this structure:
1. A struct implementing the framework interface (e.g., `BlockStorageResource`) holds `*secapi.RegionalClient`, `tenant`, `region`, and retry config.
2. `Configure()` casts `req.ProviderData.(clients)` to extract the shared clients.
3. A `*Model` struct (e.g., `BlockStorageModel`) maps Terraform schema fields using `tfsdk` tags.
4. Two private helpers convert between SDK types and model: `fooFromModel()` and `fooToModel()`.
5. Mutating operations (Create/Update) poll for `ResourceStateActive` using `GetXxxUntilState()` with configurable retry delay/interval/max_attempts.

### Type utilities (`types.go`)

Shared helpers for converting between Terraform framework types and Go types: `numberToDuration`, `numberToInt`, `toStringMap`, `fromStringMap`, `fromTime`, `fromTimePtr`, `fromRefPtr`.

### Provider configuration

Required: `token`, `tenant`, `region`, `global_providers.region_v1`. Optional: `global_providers.authorization_v1`, `retry` block (delay/interval/max_attempts in seconds). Default retry: 30s delay, 10s interval, 5 attempts.

### Linting rules

The `depguard` linter enforces use of `terraform-plugin-testing` over the older `terraform-plugin-sdk/v2` test helpers. Do not import `terraform-plugin-sdk/v2` for testing.

### Documentation generation

Docs in `docs/` are auto-generated from resource/data source schema descriptions and `examples/` `.tf` files. Run `make generate` after schema changes; CI fails if generated docs diverge from committed files.

## AI Development Scaffold

The `ai/` directory contains comprehensive documentation for AI-assisted development. Read it before implementing, reviewing, or refactoring provider code:

- `ai/README.md` — navigation and reading order
- `ai/guardrails.md` — hard rules (read first before any code change)
- `ai/implementation-checklist.md` — step-by-step checklist for new resources
- `ai/review-checklist.md` — checklist for PR reviews
- `ai/known-issues.md` — technical debt; do not accidentally work around these
- `ai/prompts/` — ready-to-use prompts for common AI tasks
