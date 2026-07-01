# Architecture

## Package Layout

```
terraform-provider-seca/
├── main.go                     # Entry point; --debug flag for Delve
├── internal/
│   ├── provider/               # All provider logic (single flat package)
│   │   ├── provider.go         # Provider struct, schema, Configure, resource/datasource lists
│   │   ├── clients.go          # Client struct, initClients(), retry constants
│   │   ├── types.go            # Shared type-conversion helpers
│   │   ├── resource_*.go       # One file per resource
│   │   ├── datasource_*.go     # One file per data source
│   │   └── *_test.go           # Unit tests for mapping functions
│   └── acctest/                # Acceptance tests (separate package)
│       ├── provider_test.go    # Shared provider factory and config builder
│       └── *_test.go           # One file per resource/data source
├── tools/                      # Separate Go module for tooling
│   ├── go.mod
│   └── tools.go                # go:generate directives
├── examples/                   # Terraform HCL examples (used by tfplugindocs)
├── docs/                       # Auto-generated documentation
└── spec/                       # Git submodule: eu-sovereign-cloud/spec
```

## Client Model

The provider initializes two SDK clients during `Configure()` and passes them to every resource and data source via `resp.DataSourceData` / `resp.ResourceData`.

```
Provider.Configure()
    └─ initClients()
           ├─ secapi.NewGlobalClient(token, endpoints)   → GlobalClient
           └─ globalClient.NewRegionalClient(ctx, region) → RegionalClient

clients struct {
    Tenant, Region string
    RetryDelay, RetryInterval time.Duration
    RetryMaxAttempts int
    GlobalClient   *secapi.GlobalClient    ← used by seca_region only
    RegionalClient *secapi.RegionalClient  ← used by all other resources/data sources
}
```

`RegionalClient` is derived from `GlobalClient` using the region name. The GlobalClient discovers the regional endpoint from the `RegionV1` global endpoint. This design means the provider must be able to reach the global region endpoint at startup; regional endpoints are resolved dynamically.

## Provider Lifecycle

```
1. Terraform reads HCL → calls Provider.Configure()
2. Configure() decodes SecaProviderModel (token, tenant, region, retry, global_providers)
3. initClients() is called:
   a. NewGlobalClient(token, {RegionV1, AuthorizationV1}) → authenticates
   b. globalClient.NewRegionalClient(ctx, region) → resolves regional endpoint
4. clients{} struct is set on resp.DataSourceData AND resp.ResourceData
5. Each resource/data source's Configure() casts ProviderData.(clients) to extract its client
```

## Resource Lifecycle

Every resource follows this exact pattern:

```
Configure()  → casts ProviderData to clients{}, stores client + tenant + region + retry params
Metadata()   → sets TypeName = providerTypeName + "_resource_name"
Schema()     → declares all attributes (see provider-conventions.md)
Create()     → reads Plan → calls CreateOrUpdateXxx() → polls GetXxxUntilState() → writes State
Read()       → reads State → calls GetXxx() → writes State
Update()     → reads Plan → calls CreateOrUpdateXxx() → polls GetXxxUntilState() → writes State
Delete()     → reads State → calls DeleteXxx()
```

Delete does **not** poll for deletion completion (see [known-issues.md](known-issues.md)).

## SDK Layering

```
Terraform Plugin Framework
    ↓ (tfsdk model structs)
internal/provider  (resource/datasource files)
    ↓ (calls)
go-sdk/secapi  (GlobalClient, RegionalClient)
    ↓ (HTTP)
go-sdk/pkg/spec/schema  (SDK types: *sdk.Workspace, *sdk.Image, etc.)
    ↓
SECA REST API
```

The mapping between Terraform models and SDK types is done exclusively by the private `xxxFromModel()` and `xxxToXxxModel()` helper functions at the bottom of each file.

## Reference System

The SECA API uses structured references to identify resources. A reference is an object with the following fields:

| Field | Required | Description | Example |
|---|---|---|---|
| `region` | no | Region; inferred from context if omitted | `eu-central-1` |
| `provider` | no | Provider+version; inferred if omitted | `seca.compute/v1` |
| `tenant` | no | Tenant; inferred if omitted | `tenant-1` |
| `workspace` | no | Workspace; inferred if omitted | `workspace-1` |
| `resource` | **yes** | Resource path within its workspace context | `instances/my-server` |

The `resource` path is `<type>/<name>` for flat resources, or `<parent-type>/<parent-name>/<child-type>/<child-name>` for hierarchical resources.

`Metadata.Ref` (stored as Terraform `id`) is the **full serialized URN** following the pattern `{provider}/{version}/tenants/{tenant}/workspaces/{workspace}/{type}/{name}`. For tenant-scoped resources the workspace segment is absent; for global resources the tenant segment is absent too. Examples by scope:

| Scope | Resource | Example `id` |
|---|---|---|
| Global | `seca_region` | `seca.region/v1/regions/region-1` |
| Tenant | `seca_workspace` | `seca.workspace/v1/tenants/tenant-1/workspaces/workspace-1` |
| Tenant | `seca_image` | `seca.storage/v1/tenants/tenant-1/images/image-1` |
| Tenant | `seca_storage_sku` | `seca.storage/v1/tenants/tenant-1/storage-skus/RD500` |
| Workspace | `seca_block_storage` | `seca.storage/v1/tenants/tenant-1/workspaces/workspace-1/block-storages/my-vol` |

**Cross-resource reference fields** (e.g., `block_storage_id` on an image, `sku_id` on a block storage) store only the `resource` path segment from the structured `sdk.Reference` object — not the full URN. The provider, tenant, and workspace are left blank in the `Reference` struct and inferred by the API from context.

## Scoping Model

SECA resources have three scopes:

| Scope | Reference type | Examples |
|---|---|---|
| Global | `sdk.GlobalResourceMetadata` | Region |
| Tenant | `sdk.RegionalResourceMetadata` (Tenant field) | Workspace, Image, StorageSku |
| Workspace | `sdk.RegionalWorkspaceResourceMetadata` (Tenant + Workspace fields) | BlockStorage |

Resources at Tenant scope use `secapi.TenantReference{Tenant, Name}` for API calls.
Resources at Workspace scope use `secapi.WorkspaceReference{Tenant, Workspace, Name}`.

## Error Handling Flow

```
API call returns error
    → resp.Diagnostics.AddError("Error <verb>ing <resource>", "..."+err.Error())
    → return

Diagnostics.Append() returns diagnostics from model mapping
    → check resp.Diagnostics.HasError() immediately after
    → return if true

State.Set() errors are appended automatically via Diagnostics.Append()
```

There is no global error handler. Every error is surfaced as a diagnostic.
