# Glossary

Domain terminology, SDK types, and Terraform model mappings.

## Domain Terms

| Term | Meaning |
|---|---|
| **SECA** | EU Sovereign Cloud (the cloud platform this provider targets) |
| **Tenant** | The top-level isolation boundary in SECA. All resources belong to a tenant. Equivalent to an "account" or "organization" in other clouds. |
| **Region** | A geographic deployment of the SECA platform. Each provider block targets exactly one region. |
| **Zone** | A sub-region availability domain within a region (e.g., zone-a, zone-b). |
| **Workspace** | A logical grouping of resources within a tenant+region. Similar to a "project" or "namespace". |
| **SKU** | Stock Keeping Unit — a named performance/feature tier for a resource type. E.g., `RD500` is a storage SKU. |
| **Block Storage** | A persistent volume that can be attached to an instance. |
| **Image** | A bootable disk image created from a block storage volume. |
| **NIC** | Network Interface Card — the network attachment point for a compute instance. |
| **Ref** | Short for "reference" — the fully qualified resource identifier, format `<kind>/<name>` (e.g., `block-storages/my-vol`). Stored as Terraform `id`. |

## SDK Types (go-sdk)

### Client Types (`secapi` package)

| Type | Purpose |
|---|---|
| `secapi.GlobalClient` | Communicates with global endpoints (Region registry, Authorization). Shared across regions. |
| `secapi.RegionalClient` | Communicates with a specific region's API endpoints. Derived from GlobalClient. |

### Reference Types (`secapi` package)

| Type | Fields | Used for |
|---|---|---|
| `secapi.TenantReference` | `Tenant TenantID`, `Name string` | Tenant-scoped resources (Workspace, Image, StorageSku) |
| `secapi.WorkspaceReference` | `Tenant TenantID`, `Workspace WorkspaceID`, `Name string` | Workspace-scoped resources (BlockStorage) |
| `secapi.TenantID` | `string` alias | Type-safe tenant identifier |
| `secapi.WorkspaceID` | `string` alias | Type-safe workspace identifier |

### Polling Type

| Type | Purpose |
|---|---|
| `secapi.ResourceObserverUntilValueConfig[T]` | Configuration for polling: `ExpectedValues []T`, `Delay`, `Interval`, `MaxAttempts` |

### Schema Types (`go-sdk/pkg/spec/schema`, aliased as `sdk`)

| Type | Terraform resource | Key fields |
|---|---|---|
| `sdk.Region` | `seca_region` data source | `Metadata *GlobalResourceMetadata`, `Spec.AvailableZones`, `Spec.Providers` |
| `sdk.Workspace` | `seca_workspace` | `Metadata *RegionalResourceMetadata`, `Status.State` |
| `sdk.Image` | `seca_image` | `Metadata *RegionalResourceMetadata`, `Spec.BlockStorageRef`, `Spec.CpuArchitecture`, `Spec.Initializer`, `Spec.Boot`, `Status.State` |
| `sdk.BlockStorage` | `seca_block_storage` | `Metadata *RegionalWorkspaceResourceMetadata`, `Spec.SizeGB`, `Spec.SkuRef`, `Spec.SourceImageRef`, `Status.State` |
| `sdk.StorageSku` | `seca_storage_sku` data source | `Metadata *SkuResourceMetadata`, `Spec.Iops`, `Spec.Type`, `Spec.MinVolumeSize` |

### Metadata Types

| Type | Scope | Fields |
|---|---|---|
| `sdk.GlobalResourceMetadata` | Global | `Name`, `Ref`, `CreatedAt`, `DeletedAt`, `LastModifiedAt` |
| `sdk.RegionalResourceMetadata` | Tenant | `Name`, `Ref`, `Tenant`, `Region`, `CreatedAt`, `DeletedAt`, `LastModifiedAt` |
| `sdk.RegionalWorkspaceResourceMetadata` | Workspace | All above + `Workspace` |
| `sdk.SkuResourceMetadata` | Tenant (SKUs) | `Name`, `Ref`, `Tenant`, `Region` (no timestamps) |

### State and Enum Types

| Type | Values | Used in |
|---|---|---|
| `sdk.ResourceState` | `ResourceStateActive`, `ResourceStateDeleted`, others | Status.State on Workspace, Image, BlockStorage |
| `sdk.ImageSpecCpuArchitecture` | `ImageSpecCpuArchitectureAmd64`, `ImageSpecCpuArchitectureArm64` | Image spec |
| `sdk.ImageSpecInitializer` | `ImageSpecInitializerNone`, `ImageSpecInitializerCloudinit22` | Image spec |
| `sdk.ImageSpecBoot` | `ImageSpecBootBIOS`, `ImageSpecBootUEFI` | Image spec |
| `sdk.StorageSkuType` | `StorageSkuTypeRemoteDurable`, others | StorageSku spec |
| `sdk.Reference` | `{Resource string}` | Cross-resource references |

## Terraform Model Types (this provider)

| Suffix | Meaning |
|---|---|
| `XxxResource` | The resource struct implementing `resource.Resource` |
| `XxxDataSource` | The data source struct implementing `datasource.DataSource` |
| `XxxModel` | The Terraform state model for a resource (tfsdk-tagged struct) |
| `XxxDataSourceModel` | The Terraform state model for a data source |
| `xxxFromModel()` | Converts resource Terraform model → SDK type (for Create/Update) |
| `xxxToResourceModel()` | Converts SDK type → resource Terraform model |
| `xxxToDataSourceModel()` | Converts SDK type → data source Terraform model |

## Infrastructure Terms

| Term | Meaning in this codebase |
|---|---|
| `global_providers` | Provider config block pointing to the global API endpoints (region registry, authorization) |
| `region_v1` | URL of the SECA Region API v1 global endpoint |
| `authorization_v1` | URL of the SECA Authorization API v1 global endpoint |
| `retry` | Provider config block controlling polling behavior for async resources |
