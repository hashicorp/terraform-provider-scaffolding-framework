# Terraform Plugin Framework Usage

This document captures how the Terraform Plugin Framework (`terraform-plugin-framework v1.19.0`) is used in this provider. It is not a rehash of the official docs — it focuses on decisions and patterns specific to this codebase.

## Framework Version and Protocol

- Framework: `github.com/hashicorp/terraform-plugin-framework v1.19.0`
- Protocol version: 6 (via `providerserver.NewProtocol6WithError`)
- Provider registration: `providerserver.Serve()` in `main.go` with `Address: "registry.terraform.io/eu-sovereign-cloud/seca"`

## Schema Types in Use

| Terraform type | Framework type | Used for |
|---|---|---|
| String | `tfschema.StringAttribute` | names, IDs, timestamps, enums |
| Int64 | `tfschema.Int64Attribute` | sizes (GB), counts (IOPS) |
| Map(String) | `tfschema.MapAttribute{ElementType: types.StringType}` | labels, annotations, extensions |
| List(String) | `tfschema.ListAttribute{ElementType: types.StringType}` | available_zones |
| List(Object) | `tfschema.ListNestedAttribute` | providers list in region |
| Number | `schema.NumberAttribute` | retry seconds in provider config |

`Number` is used only in the provider schema for retry config because seconds need fractional support (`1.5` = 1500ms). Resource schemas use `Int64` for integer quantities.

## Attribute Classification

| Classification | When to use |
|---|---|
| `Required` | Must be specified by user; drives API calls |
| `Computed` | Set by the API or derived from provider config |
| `Optional + Computed` | User may set it; API may default or modify it |

Current `Optional + Computed` attributes:
- `seca_image.initializer` — user may set; API defaults if omitted
- `seca_image.boot` — user may set; API defaults if omitted

## Plan Modifiers

Only `stringplanmodifier.RequiresReplace()` is used. It is applied to:
- `name` on all resources (SECA names are immutable)
- `workspace_id` on workspace-scoped resources

No `UseStateForUnknown()` is currently used. This means computed fields show `(known after apply)` on every plan even when unchanged — a known gap.

## Reading State vs Plan vs Config

| Operation | Where to read input from |
|---|---|
| Create | `req.Plan.Get()` |
| Read | `req.State.Get()` |
| Update | `req.Plan.Get()` |
| Delete | `req.State.Get()` |
| Data source Read | `req.Config.Get()` |

This pattern is consistent across all resources.

## Null and Unknown Handling

The framework distinguishes three states for any attribute value: set, null, unknown.

For Optional fields that may not be sent to the API:
```go
if !data.SourceImageId.IsNull() && !data.SourceImageId.IsUnknown() {
    block.Spec.SourceImageRef = &sdk.Reference{Resource: data.SourceImageId.ValueString()}
}
```

For `fromRefPtr()` — an optional pointer-to-Reference returned by the API:
```go
func fromRefPtr(ref *sdk.Reference) types.String {
    if ref == nil {
        return types.StringNull()
    }
    return types.StringValue(ref.Resource)
}
```

The `types_test.go` tests explicitly verify null/unknown/zero behavior for all helpers.

## Diagnostics

Use `resp.Diagnostics.Append(...)` to collect errors from sub-operations:
```go
data, diags := xxxToModel(ctx, sdkObj)
resp.Diagnostics.Append(diags...)
if resp.Diagnostics.HasError() {
    return
}
```

Check `HasError()` immediately after every `Append()` or `Get()` call. Do not continue processing after an error is added.

## Nested Objects (Lists of Objects)

The region data source demonstrates the pattern for `ListNestedAttribute`:

```go
// Schema
"providers": tfschema.ListNestedAttribute{
    Computed: true,
    NestedObject: tfschema.NestedAttributeObject{
        Attributes: map[string]tfschema.Attribute{...},
    },
},

// Model struct
type RegionProviderModel struct {
    Name    types.String `tfsdk:"name"`
    ...
}

// Attr types map (required for ListValueFrom)
var regionProviderAttrTypes = map[string]attr.Type{
    "name":    types.StringType,
    ...
}

// Mapping
providers := make([]RegionProviderModel, 0, len(region.Spec.Providers))
for _, p := range region.Spec.Providers {
    providers = append(providers, RegionProviderModel{...})
}
providersList, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: regionProviderAttrTypes}, providers)
```

Always pre-declare the `AttrTypes` map as a package-level `var` for reuse in tests.

## Features Not Yet Used

The following framework features are available but not yet used in this provider. Do not introduce them without first discussing with the team and updating this document:

| Feature | Reason not yet used |
|---|---|
| `ConfigValidators` | No cross-field validation needed yet |
| `Validators` (field-level) | No format validation implemented |
| `UseStateForUnknown()` | Known gap; all computed fields re-render on plan |
| `ImportState` | Not implemented; see [known-issues.md](known-issues.md) |
| `StateUpgraders` | No schema version changes yet |
| `tflog` | No structured logging implemented |
| Timeouts (`timeouts` block) | Not implemented; retry config is used instead |
| `SemanticEquals` | Not implemented |
| `Identity` | Not implemented |

When implementing these, reference the official framework documentation and follow patterns from existing resources.
