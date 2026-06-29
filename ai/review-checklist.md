# Review Checklist

Use this checklist when reviewing any PR that adds or modifies provider code.

## Architecture

- [ ] New files follow the naming convention (`resource_xxx.go`, `datasource_xxx.go`)
- [ ] All provider code is in `internal/provider/` (flat, no sub-packages)
- [ ] New resources are registered in `provider.go` `Resources()` list
- [ ] New data sources are registered in `provider.go` `DataSources()` list
- [ ] Acceptance tests are in `internal/acctest/` under `package acctest`

## Terraform Plugin Framework

- [ ] Compile-time interface checks are present (`var _ resource.Resource = (*XxxResource)(nil)`)
- [ ] `Configure()` checks for `req.ProviderData == nil` before type assertion
- [ ] `Configure()` type-asserts to `clients` (not a pointer) and reports error on failure
- [ ] Create reads from `req.Plan`, not `req.State`
- [ ] Read reads from `req.State`, not `req.Plan`
- [ ] Update reads from `req.Plan`, not `req.State`
- [ ] Delete reads from `req.State`, not `req.Plan`
- [ ] Data source Read reads from `req.Config`, not `req.State`
- [ ] `resp.State.Set(ctx, &data)` uses a pointer and its diagnostics are appended

## Schema Design

- [ ] Standard attributes present: `id` (Computed), `name` (Required + RequiresReplace), `tenant` (Computed), `region` (Computed), `created_at` (Computed), `deleted_at` (Computed), `last_modified_at` (Computed)
- [ ] `labels`, `annotations`, `extensions` are `Optional` on resources, `Computed` on data sources
- [ ] `workspace_id` on workspace-scoped resources is `Required + RequiresReplace`
- [ ] `id` is always `Computed: true`
- [ ] No `RequiresReplace()` on Computed attributes
- [ ] `RequiresReplace()` is only used for attributes the API does not support in-place updates for
- [ ] Data source schemas include a `state` field for resources with a status

## Async Operations

- [ ] Create calls `GetXxxUntilState()` after `CreateOrUpdateXxx()`
- [ ] Update calls `GetXxxUntilState()` after `CreateOrUpdateXxx()`
- [ ] State is populated from the **polling result**, not the initial create/update response
- [ ] The `ResourceObserverUntilValueConfig` uses the resource struct's retry fields (not hard-coded values)
- [ ] The reference for polling is built from the API result, not the Terraform model
- [ ] Both the initial call error and the polling error are handled and surfaced as diagnostics

## Error Handling

- [ ] Every SDK call error results in `resp.Diagnostics.AddError()` + `return`
- [ ] `resp.Diagnostics.HasError()` is checked after every `Diagnostics.Append()`
- [ ] No error is silently discarded (`_ = err` is absent)
- [ ] Error summary follows: `"Error <verb>ing <resource type>"`
- [ ] Error detail follows: `"An error was encountered when <verb>ing...\nError: "+err.Error()`
- [ ] Polling error detail: `"...while waiting for the <resource type> to become active.\nError: "+err.Error()`

## Mapping Functions

- [ ] `xxxFromModel(tenant string, data XxxModel)` receives tenant as parameter (not from model)
- [ ] `xxxToResourceModel` and `xxxToDataSourceModel` return `diag.Diagnostics`
- [ ] All type conversion uses helpers from `types.go`, not inline conversions
- [ ] Optional pointer fields use `IsNull() && IsUnknown()` guard before setting
- [ ] `id` is set from `Metadata.Ref`, not from `Metadata.Name`
- [ ] SDK enum fields are cast: `sdk.SomeEnum(data.Field.ValueString())`
- [ ] Resource model does NOT include `state` (data source model does)
- [ ] Data source model mappers do NOT inline resource model mappers — they are separate functions

## State Management

- [ ] No backwards-incompatible schema changes (removed/renamed/retyped attributes)
- [ ] If schema layout changed: state upgrader is implemented
- [ ] `deleted_at` uses `fromTimePtr()` (nullable) not `fromTime()`

## Go Code Quality

- [ ] `gofumpt` formatting (run `make fmt` to verify)
- [ ] No `terraform-plugin-sdk/v2` imports
- [ ] `context.Context` is passed to every SDK call (not `context.Background()`)
- [ ] No `time.Sleep()` calls
- [ ] No `panic()` calls
- [ ] Linter passes (`make lint`)

## Tests

- [ ] Unit test for `xxxToResourceModel` present and covers all fields
- [ ] Unit test for `xxxToDataSourceModel` present and covers all fields including status
- [ ] Unit tests cover null/zero/empty cases for all nullable fields
- [ ] Acceptance test present in `internal/acctest/`
- [ ] Acceptance test uses `testAccProtoV6ProviderFactories` and `testAccPreCheck`
- [ ] Acceptance test uses `resource.ComposeAggregateTestCheckFunc`
- [ ] No `terraform-plugin-sdk/v2` test helpers imported

## Documentation

- [ ] `make generate` was run after schema changes
- [ ] Docs in `docs/` match the current schema
- [ ] Example `.tf` files in `examples/` are valid and formatted (`terraform fmt`)
