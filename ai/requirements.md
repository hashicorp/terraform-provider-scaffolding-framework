# Requirements

## Version Matrix

| Component | Version |
|---|---|
| Go | `1.26.4` (see `go.mod`) |
| Terraform Plugin Framework | `v1.19.0` |
| Terraform Plugin Testing | `v1.16.0` |
| go-sdk | `v0.4.2` |
| Supported Terraform CLI | `1.13.*`, `1.14.*` (CI matrix in `.github/workflows/test.yml`) |
| Registry address | `registry.terraform.io/eu-sovereign-cloud/seca` |

**Do not use `terraform-plugin-sdk/v2`** for any new code. The `depguard` linter enforces this.

## Provider Goals

The SECA provider exposes the EU Sovereign Cloud (SECA) platform as Terraform-managed infrastructure. It wraps the `go-sdk` and presents SECA concepts as Terraform resources and data sources, handling:

- Authentication via bearer token
- Regional resource routing through the GlobalClient → RegionalClient bootstrap
- Async provisioning — every mutating API call is eventually consistent; the provider polls until resources reach `Active` state

## Provider Non-Goals

- No multi-region resources within a single provider block (each provider block targets exactly one `region`)
- No sub-tenant isolation — `tenant` comes from provider config and is immutable per provider instance
- No cross-tenant references — all `tenant` fields in resource models are read-only (computed from provider config)
- No Terraform state migration compatibility between major versions of the SECA API

## Supported Resources

| Resource | API client | Scope |
|---|---|---|
| `seca_workspace` | `RegionalClient.WorkspaceV1` | Tenant |
| `seca_image` | `RegionalClient.StorageV1` | Tenant |
| `seca_block_storage` | `RegionalClient.StorageV1` | Workspace |

## Supported Data Sources

| Data Source | API client | Scope |
|---|---|---|
| `seca_region` | `GlobalClient.RegionV1` | Global |
| `seca_workspace` | `RegionalClient.WorkspaceV1` | Tenant |
| `seca_image` | `RegionalClient.StorageV1` | Tenant |
| `seca_block_storage` | `RegionalClient.StorageV1` | Workspace |
| `seca_storage_sku` | `RegionalClient.StorageV1` | Tenant |

## Planned but Not Yet Implemented

See [roadmap.md](roadmap.md) for the full list.

## Backward Compatibility Policy

- Schema attributes that are `Required` or `Optional` must never be removed or have their type changed in a minor release without a state migration.
- Computed attributes may be added freely (they are additive and do not break existing configurations).
- `name` is used as the resource identifier in the SECA API and is always `RequiresReplace` — renaming a resource destroys and recreates it.
- All API resource references use the `<kind>/<name>` format (e.g., `block-storages/my-vol`). This format must be preserved in Terraform `id` values.

## Compatibility Guarantees

- The provider targets Terraform `>= 1.13`.
- Protocol version is 6 (`tfprotov6`), served via `providerserver.NewProtocol6WithError`.
- Binary is built with `CGO_ENABLED=0` for static linking.
