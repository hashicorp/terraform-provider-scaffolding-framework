# Implementation Checklist

Use this checklist when implementing a new resource or data source. Complete every step in order.

## Before Writing Code

- [ ] Identify the resource scope: Global, Tenant, or Workspace (see [architecture.md](architecture.md#scoping-model))
- [ ] Identify the SDK client: `GlobalClient` or `RegionalClient`
- [ ] Identify the SDK service: e.g., `WorkspaceV1`, `StorageV1`
- [ ] Identify the SDK methods: `CreateOrUpdateXxx`, `GetXxx`, `GetXxxUntilState`, `DeleteXxx`
- [ ] Identify the SDK type: `*sdk.Xxx`, its `Metadata` type, and its `Spec`/`Status` types
- [ ] List all Required, Optional, and Computed fields from the API spec
- [ ] Determine which fields are immutable (will need `RequiresReplace()`)
- [ ] Determine which fields the API may default or modify (will need `Optional + Computed`)
- [ ] Read [guardrails.md](guardrails.md) before proceeding

## Resource File (`internal/provider/resource_xxx.go`)

- [ ] Create file named `resource_<name>.go`
- [ ] Declare compile-time interface checks
- [ ] Define `XxxResource` struct with: `client`, `tenant`, `region`, `retryDelay`, `retryInterval`, `retryMaxAttempts`
- [ ] Implement `newXxxResource() resource.Resource`
- [ ] Implement `Metadata()` — set TypeName = `providerTypeName + "_xxx"`
- [ ] Define `XxxModel` struct with all fields (standard + resource-specific), in the correct order
- [ ] Implement `Schema()` — include all standard attributes; add resource-specific attributes
- [ ] Apply `RequiresReplace()` to `name` and `workspace_id`
- [ ] Implement `Configure()` — guard nil check, type assert to `clients`, store all fields including retry
- [ ] Implement `Create()`:
  - Read from `req.Plan`
  - Call `xxxFromModel(tenant, data)` to build SDK object
  - Call `CreateOrUpdateXxx()`
  - Build reference for polling (from API result, not model)
  - Build `ResourceObserverUntilValueConfig` with struct retry fields
  - Call `GetXxxUntilState()`
  - Call `xxxToResourceModel(ctx, result)`
  - Write state
- [ ] Implement `Read()`:
  - Read from `req.State`
  - Build reference from state values
  - Call `GetXxx()`
  - Call `xxxToResourceModel(ctx, result)`
  - Write state
- [ ] Implement `Update()` — same structure as Create
- [ ] Implement `Delete()`:
  - Read from `req.State`
  - Build minimal SDK object (Metadata only)
  - Call `DeleteXxx()`
- [ ] Implement `xxxFromModel(tenant string, data XxxModel) *sdk.Xxx`
- [ ] Implement `xxxToResourceModel(ctx, *sdk.Xxx) (XxxModel, diag.Diagnostics)`

## Data Source File (`internal/provider/datasource_xxx.go`)

- [ ] Create file named `datasource_<name>.go`
- [ ] Declare compile-time interface checks
- [ ] Define `XxxDataSource` struct with: `client`, `tenant` (and `region` if needed)
- [ ] Implement `newXxxDataSource() datasource.DataSource`
- [ ] Implement `Metadata()` — set TypeName = `providerTypeName + "_xxx"`
- [ ] Define `XxxDataSourceModel` struct (same fields as resource model + `state`)
- [ ] Implement `Schema()` — `labels`/`annotations`/`extensions` are `Computed` (not Optional); add `state`
- [ ] Implement `Configure()` — guard nil check, type assert, store client and tenant
- [ ] Implement `Read()`:
  - Read from `req.Config`
  - Build reference from config values
  - Call `GetXxx()`
  - Call `xxxToDataSourceModel(ctx, result)`
  - Write state
- [ ] Implement `xxxToDataSourceModel(ctx, *sdk.Xxx) (XxxDataSourceModel, diag.Diagnostics)` — include `State` from status

## Registration

- [ ] Add `newXxxResource` to `Resources()` list in `provider.go`
- [ ] Add `newXxxDataSource` to `DataSources()` list in `provider.go`

## Unit Tests (`internal/provider/resource_xxx_test.go`)

- [ ] Create file named `resource_<name>_test.go` for `TestXxxToResourceModel`
- [ ] Create file named `datasource_<name>_test.go` for `TestXxxToDataSourceModel`
- [ ] Build a fully-populated SDK object
- [ ] Include `DeletedAt` as non-nil (tests nullable pointer)
- [ ] Include non-empty `Labels`, `Annotations`, `Extensions`
- [ ] Assert every field in the resulting model
- [ ] Assert `IsNull()` for absent optional pointer fields (separate test case)
- [ ] Assert the `state` field for data source tests

## Acceptance Tests (`internal/acctest/<name>_test.go`)

- [ ] Create file named `<name>_test.go` in `internal/acctest/`
- [ ] Define `testAcc<Name>ResourceConfig() string`
- [ ] Define `testAcc<Name>DataSourceConfig() string`
- [ ] Implement `TestAcc<Name>` with at minimum:
  - Step 1: Create resource; check all non-computed fields are set correctly; check tenant and region
  - Step 2: Create resource + data source; check data source reads resource; check `state = "active"`

## Documentation

- [ ] Add example HCL in `examples/resources/seca_<name>/resource.tf`
- [ ] Add example HCL in `examples/data-sources/seca_<name>/data-source.tf`
- [ ] Run `make generate` to regenerate `docs/`
- [ ] Verify generated docs look correct

## Before Opening PR

The CI `generate` job will fail if `docs/` was not regenerated. **Do not skip this.**

- [ ] `make generate` — run and commit the output (including any `docs/` changes)
- [ ] `git diff --exit-code docs/` — must produce no output (clean)
- [ ] `make build` — must succeed
- [ ] `make test` — all tests pass
- [ ] `make lint` — no linter errors
