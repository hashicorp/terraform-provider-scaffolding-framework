# Provider Conventions

This document captures every naming, schema, and structural convention observed across the existing implementation. New resources and data sources must follow these conventions exactly.

## File Naming

| Type | File name pattern |
|---|---|
| Resource | `resource_<name>.go` |
| Data source | `datasource_<name>.go` |
| Resource unit test | `resource_<name>_test.go` |
| Data source unit test | `datasource_<name>_test.go` |

Acceptance tests live in `internal/acctest/` as `<name>_test.go`.

## Type Naming

| Purpose | Naming pattern | Example |
|---|---|---|
| Resource struct | `<Name>Resource` | `BlockStorageResource` |
| Data source struct | `<Name>DataSource` | `BlockStorageDataSource` |
| Resource model | `<Name>Model` | `BlockStorageModel` |
| Data source model | `<Name>DataSourceModel` | `BlockStorageDataSourceModel` |
| Constructor (resource) | `new<Name>Resource()` | `newBlockStorageResource()` |
| Constructor (data source) | `new<Name>DataSource()` | `newBlockStorageDataSource()` |
| Model→SDK mapper | `<name>FromModel(tenant, data)` | `blockStorageFromModel()` |
| SDK→resource model | `<name>ToResourceModel(ctx, sdk)` | `blockStorageToResourceModel()` |
| SDK→data source model | `<name>ToDataSourceModel(ctx, sdk)` | `blockStorageToDataSourceModel()` |

## Interface Compliance

Every resource and data source file must declare compile-time interface checks:

```go
var (
    _ resource.Resource              = (*BlockStorageResource)(nil)
    _ resource.ResourceWithConfigure = (*BlockStorageResource)(nil)
)
```

For data sources:
```go
var (
    _ datasource.DataSource              = (*BlockStorageDataSource)(nil)
    _ datasource.DataSourceWithConfigure = (*BlockStorageDataSource)(nil)
)
```

## TypeName Convention

```go
func (r *BlockStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_block_storage"
}
```

The suffix is always snake_case matching the Terraform resource name.

## Standard Schema Attributes

Every resource schema must include these attributes in this order:

```go
"id":               Computed: true  (populated from Metadata.Ref)
"name":             Required: true + RequiresReplace()  (immutable API identifier)
"tenant":           Computed: true  (from provider config, read-only)
"region":           Computed: true  (from provider config, read-only)
"created_at":       Computed: true  (RFC3339 string)
"deleted_at":       Computed: true  (RFC3339 string, nullable)
"last_modified_at": Computed: true  (RFC3339 string)
"labels":           Optional: true  + MapAttribute{ElementType: types.StringType}
"annotations":      Optional: true  + MapAttribute{ElementType: types.StringType}
"extensions":       Optional: true  + MapAttribute{ElementType: types.StringType}
```

Workspace-scoped resources add:
```go
"workspace_id": Required: true + RequiresReplace()
```

Data source schemas follow the same pattern with these differences:
- `labels`, `annotations`, `extensions` are `Computed: true` (not Optional)
- Resource-specific status fields (e.g., `state`) are added as `Computed: true`
- `workspace_id` on data sources is `Required: true` without `RequiresReplace()`

**Exception — SKU-type schemas backed by `SkuResourceMetadata`:** SKU catalog
data sources (e.g., `seca_storage_sku`) are backed by `sdk.SkuResourceMetadata`,
which has no timestamp fields. These schemas intentionally omit `created_at`,
`deleted_at`, and `last_modified_at` — they expose only `id`, `name`, `tenant`,
`region`, plus their spec fields. Do not add timestamp attributes to SKU schemas;
the underlying metadata cannot populate them. See [glossary.md](glossary.md) for
the `SkuResourceMetadata` field list.

## ForceNew / RequiresReplace Rules

`stringplanmodifier.RequiresReplace()` is applied when changing the field would require destroying and recreating the resource at the API level:

- `name` — always RequiresReplace (SECA resource names are immutable identifiers)
- `workspace_id` — always RequiresReplace (resources cannot move workspaces)
- Other immutable spec fields — add RequiresReplace with a comment explaining the API constraint

Never add RequiresReplace to Computed fields or to fields that the API supports in-place updates for.

## Model Field Ordering

Within a model struct, follow this ordering convention:
1. `Id`
2. `Name`
3. `WorkspaceId` (if workspace-scoped)
4. `Tenant`, `Region`
5. `CreatedAt`, `DeletedAt`, `LastModifiedAt`
6. `Labels`, `Annotations`, `Extensions`
7. Resource-specific spec fields
8. Status fields (data sources only)

## Configure() Pattern

Every resource and data source Configure() must:

```go
func (r *XxxResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return  // called before provider.Configure(); safe to ignore
    }

    clients, ok := req.ProviderData.(clients)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected provider data type",
            fmt.Sprintf("Expected sdk.Clients, got: %T", req.ProviderData),
        )
        return
    }

    r.client = clients.RegionalClient  // or GlobalClient for global resources
    r.tenant = clients.Tenant
    r.region = clients.Region
    r.retry = retryConfig{
        delay:       clients.RetryDelay,
        interval:    clients.RetryInterval,
        maxAttempts: clients.RetryMaxAttempts,
    }
}
```

`r.retry` (a `retryConfig`, defined in `retry.go`) holds the resolved provider-level retry values. Each resource also exposes an optional per-resource `retry` block — add `Retry *SecaRetryModel` to the model, declare it in the schema via `retryResourceSchema()`, and resolve the effective config at poll time with `r.retry.with(data.Retry)` (see [async-operations.md](async-operations.md)).

Data sources that do not perform async operations do not need the retry field or block.

## Diagnostic Error Messages

Error messages follow a two-part convention:
- **Summary**: `"Error <verb>ing <resource type>"` — e.g., `"Error creating block storage"`
- **Detail**: `"An error was encountered when <verb>ing the <resource type>.\nError: " + err.Error()`

For polling errors, the summary uses the **read** verb (the polling step is a `GetXxxUntilState` call, not the mutation), regardless of whether the caller is `Create()` or `Update()`:
- **Summary**: `"Error reading <resource type>"` — e.g., `"Error reading block storage"`
- **Detail**: `"An error was encountered while waiting for the <resource type> to become active.\nError: " + err.Error()`

## id Field Value

`id` is always set to `Metadata.Ref`, which is the full `<kind>/<name>` reference string returned by the API (e.g., `"block-storages/my-vol"`). Never set `id` to just the name.

## Tenant Handling

Tenant is sourced from provider config and passed to resources via `clients.Tenant`. It is:
- Never read from the Terraform config by resources (it is always Computed)
- Always injected into the `Metadata.Tenant` field when building SDK objects in `xxxFromModel()`
- Always passed as the first argument to `xxxFromModel(tenant string, data Model)`

## Import

All resources implement `resource.ResourceWithImportState`. The import ID must carry exactly the identity fields `Read()` looks up — the tenant always comes from provider config, never from the import ID:

- **Tenant-scoped** (`seca_workspace`, `seca_image`): the ID is the resource `name`, set via `resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)`.
- **Workspace-scoped** (`seca_block_storage`): the ID is a composite `<workspace_id>/<name>`, parsed with `strings.Cut` and written with `resp.State.SetAttribute()` for each part (with a clear diagnostic if the format is wrong).

Do **not** passthrough to `id`: `Read()` keys off `name`/`workspace_id`, not the `Metadata.Ref` stored in `id` (see [id Field Value](#id-field-value)). Name the `ImportState` receiver `r`, not `resource`, so the body can reach `resource.ImportStatePassthroughID` (the usual `resource` receiver name shadows the package).

### Documenting the import ID format

Import docs are generated by `tfplugindocs` (via `make generate`) — never hand-edit `docs/`. Each resource's import ID format is documented by two example files that the generator renders into the resource's **## Import** section:

- `examples/resources/<type>/import.sh` — the CLI `terraform import` command
- `examples/resources/<type>/import-by-string-id.tf` — the config-driven `import {}` block with the `id` attribute (Terraform ≥ 1.5). Note the exact filename: `tfplugindocs` ignores a plain `import.tf`.

Keep the IDs in those examples in sync with the `ImportState` implementation.
