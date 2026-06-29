# Roadmap

This document captures planned resources and features. It is derived from the `examples/use-cases/usage.tf` file, which shows the intended full usage of the provider, and from the `examples/data-sources/` and `examples/resources/` directories.

## Currently Implemented

| Type | Name | API Service |
|---|---|---|
| Resource | `seca_workspace` | `WorkspaceV1` |
| Resource | `seca_image` | `StorageV1` |
| Resource | `seca_block_storage` | `StorageV1` |
| Data source | `seca_region` | `RegionV1` (global) |
| Data source | `seca_workspace` | `WorkspaceV1` |
| Data source | `seca_image` | `StorageV1` |
| Data source | `seca_block_storage` | `StorageV1` |
| Data source | `seca_storage_sku` | `StorageV1` |

## Planned Resources

These resources appear in `examples/` and `examples/use-cases/usage.tf` but are not yet implemented. They represent the intended full scope of the provider.

### Compute

| Type | Name | Notes |
|---|---|---|
| Resource | `seca_instance` | Has `boot_volume`, `primary_nic_id`, `ssh_keys`, `zone`, `sku_id` |
| Data source | `seca_instance` | Read-only view of an instance |
| Data source | `seca_instance_sku` | Discover available instance SKUs |

### Networking

| Type | Name | Notes |
|---|---|---|
| Resource | `seca_network` | Has `sku_id`, `cidr.ipv4` |
| Resource | `seca_subnet` | Has `network_id`, `cidr.ipv4`, `route_table_id` |
| Resource | `seca_route_table` | Has `network_id`, `routes` list (destination + target) |
| Resource | `seca_internet_gateway` | Workspace-scoped |
| Resource | `seca_security_group` | Has `rules` list with direction, protocol, ports, source_refs |
| Resource | `seca_public_ip` | Has `version` (IPv4/IPv6) |
| Resource | `seca_nic` | Has `subnet_id`, `addresses`, `public_ip_id` |
| Data source | `seca_network` | Read-only view |
| Data source | `seca_network_sku` | Discover available network SKUs |
| Data source | `seca_internet_gateway` | Read-only view |
| Data source | `seca_security_group` | Read-only view |
| Data source | `seca_subnet` | Read-only view |
| Data source | `seca_route_table` | Read-only view |
| Data source | `seca_public_ip` | Read-only view |
| Data source | `seca_nic` | Read-only view |

### Authorization

| Type | Name | Notes |
|---|---|---|
| Resource | `seca_role` | Uses `GlobalClient.AuthorizationV1` |
| Resource | `seca_role_assignment` | Uses `GlobalClient.AuthorizationV1` |
| Data source | `seca_role` | Read-only view |
| Data source | `seca_role_assignment` | Read-only view |

### Infrastructure Debt (Planned Improvements)

See [known-issues.md](known-issues.md) for the full list. Priority items:

1. **ImportState** — all existing resources need it before the provider can be considered production-ready
2. **404 handling in Read** — necessary for drift detection and out-of-band deletions
3. **Delete polling** — necessary to prevent race conditions on recreate
4. **Acceptance test endpoints from env vars** — necessary for CI/CD against different environments
5. **`UseStateForUnknown()`** — quality of life improvement for Terraform plans

## Implementation Priority Notes

When implementing networking resources, note that:
- `seca_network` must exist before `seca_subnet`, `seca_route_table`
- `seca_internet_gateway` is referenced as a route target
- `seca_security_group.rules` uses a nested list of objects — reference the `seca_region.providers` pattern in `datasource_region.go`

Authorization resources use `GlobalClient.AuthorizationV1`, not `RegionalClient`. The `RegionDataSource` is the only current example of using `GlobalClient`.
