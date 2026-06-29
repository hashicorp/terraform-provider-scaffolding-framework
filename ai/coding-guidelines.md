# Coding Guidelines

## Formatting

- Formatter: `gofumpt` (stricter than `gofmt`; run via `make fmt`)
- Linter: `golangci-lint v2` (run via `make lint`)
- CI will reject builds that do not pass both

## Imports

Import alias conventions observed in the codebase:

```go
import (
    // stdlib first (no alias)
    "context"
    "fmt"
    "time"

    // framework packages (use tfschema alias where needed to avoid collision)
    "github.com/hashicorp/terraform-plugin-framework/datasource"
    tfschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"

    // SDK packages (always alias spec/schema as sdk)
    sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
    "github.com/eu-sovereign-cloud/go-sdk/secapi"
)
```

The `sdk` alias is always used for `go-sdk/pkg/spec/schema`. The `tfschema` alias is used for both `datasource/schema` and `resource/schema` packages to prevent name collision with the `schema` sub-package within the same file.

## Package Structure

- All provider code lives in `package provider` (`internal/provider/`)
- All acceptance tests live in `package acctest` (`internal/acctest/`)
- Unit tests in `internal/provider/` use `package provider` (white-box)
- Never create sub-packages inside `internal/provider/` — all code stays flat

## Naming

- Resource types: `XxxResource` (PascalCase, singular noun)
- Model types: `XxxModel` (for resources), `XxxDataSourceModel` (for data sources)
- Private helpers: `xxxFromModel`, `xxxToResourceModel`, `xxxToDataSourceModel` (camelCase)
- Constructor functions: `newXxxResource()`, `newXxxDataSource()`
- Avoid abbreviations unless they are domain terms (e.g., `sku`, `nic`)

## Error Handling

- **Always** check errors from SDK calls immediately
- **Always** call `resp.Diagnostics.AddError()` before returning on error
- **Always** check `resp.Diagnostics.HasError()` after `Diagnostics.Append()`
- Do **not** wrap errors redundantly — the SDK error message is already included in the detail string
- Do **not** use `panic()` — use diagnostics

```go
// Correct
result, err := r.client.StorageV1.CreateOrUpdateBlockStorage(ctx, block)
if err != nil {
    resp.Diagnostics.AddError(
        "Error creating block storage",
        "An error was encountered when creating the block storage.\nError: "+err.Error(),
    )
    return
}

// Wrong — ignoring err
result, _ := r.client.StorageV1.CreateOrUpdateBlockStorage(ctx, block)
```

## Context

- Always pass `ctx` as the first argument to SDK calls
- Never create a new `context.Background()` inside a resource method — use the one provided
- The framework manages cancellation; respect it by passing ctx everywhere

## Mapping Functions

Each resource/data source file defines private mapping functions at the bottom:

```go
// SDK type → resource Terraform model
func xxxToResourceModel(ctx context.Context, obj *sdk.Xxx) (XxxModel, diag.Diagnostics)

// SDK type → data source Terraform model
func xxxToDataSourceModel(ctx context.Context, obj *sdk.Xxx) (XxxDataSourceModel, diag.Diagnostics)

// Resource Terraform model → SDK type  (Create/Update only)
func xxxFromModel(tenant string, data XxxModel) *sdk.Xxx
```

Rules for mapping functions:
- `xxxFromModel` receives `tenant string` as first argument (never read tenant from the model)
- `xxxToResourceModel` and `xxxToDataSourceModel` return `diag.Diagnostics` even if currently empty — future-proofing
- Use shared helpers from `types.go` — do not re-implement `fromTime`, `fromRefPtr`, etc.
- Data source models include `state` from `sdk.Status.State`; resource models do not (state is an API-side concept)

## Type Conversion Helpers (`types.go`)

Always use these helpers for type conversion. Never convert inline:

| Helper | Input → Output | Null behavior |
|---|---|---|
| `fromTime(t)` | `time.Time` → `types.String` | zero time → `types.StringNull()` |
| `fromTimePtr(t)` | `*time.Time` → `types.String` | nil or zero → `types.StringNull()` |
| `fromRefPtr(r)` | `*sdk.Reference` → `types.String` | nil → `types.StringNull()` |
| `fromStringMap(ctx, m)` | `map[string]string` → `(types.Map, diag.Diagnostics)` | nil/empty → null map |
| `toStringMap(m)` | `types.Map` → `map[string]string` | null/unknown → nil |
| `numberToDuration(n)` | `types.Number` → `time.Duration` | null/unknown → 0 |
| `numberToInt(n)` | `types.Number` → `int` | null/unknown → 0 |

## SDK Enum Casting

SDK enums are string types. Cast them explicitly:

```go
// To SDK type
sdk.ImageSpecCpuArchitecture(data.CpuArchitecture.ValueString())

// From SDK type
types.StringValue(string(image.Spec.CpuArchitecture))
```

## Unused Variables / Assignments

The linter enforces `unused`, `ineffassign`. Do not assign to `_` to suppress; fix the root cause.

## Comments

- No multi-line or multi-paragraph comments
- One-line comments only, and only when the WHY is non-obvious
- The existing `// Create the block storage`, `// Wait until it is active` comments are structural separators — acceptable but do not add new ones
- Do not add `// TODO` comments without a linked issue

## Dependencies

- Never add a new dependency without approval
- Never import `terraform-plugin-sdk/v2` — the depguard linter blocks this
- Use `go.mod`; never use `replace` directives for published modules
