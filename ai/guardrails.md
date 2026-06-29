# AI Guardrails

These are hard rules. Violations will cause bugs, regressions, or data loss.

## Schema

**Never:**
- Remove, rename, or change the type of an existing Required or Optional attribute — this is a breaking change for existing Terraform configurations
- Add a `RequiresReplace()` plan modifier to an attribute that the API supports in-place updates for
- Change `id` to anything other than `Metadata.Ref` (the full `<kind>/<name>` reference string)
- Add `RequiresReplace()` to Computed attributes
- Make a Computed attribute Optional+Computed without understanding whether the API will echo back the value or default it

**Always:**
- Maintain the standard attribute set (`id`, `name`, `tenant`, `region`, `created_at`, `deleted_at`, `last_modified_at`, `labels`, `annotations`, `extensions`) on every resource and data source that represents a regional resource
- Apply `RequiresReplace()` to `name` on every resource
- Apply `RequiresReplace()` to `workspace_id` on every workspace-scoped resource

## Async Operations

**Never:**
- Write Terraform state from the result of `CreateOrUpdateXxx()` — write from the result of `GetXxxUntilState()` only
- Skip the polling step for Create or Update
- Use `time.Sleep()` instead of the SDK's `GetXxxUntilState()` mechanism
- Hard-code delay or interval values — always read from resource struct retry fields

**Always:**
- Poll for `ResourceStateActive` after every Create and Update
- Use the retry parameters from the `clients` struct (passed through from provider config)
- Check for error from `GetXxxUntilState()` and surface it as a diagnostic

## Error Handling

**Never:**
- Ignore an error from an SDK call (including during polling)
- Use `_` to discard an error return value from an API call
- Continue processing after `resp.Diagnostics.HasError()` returns true
- Panic

**Always:**
- Call `resp.Diagnostics.AddError()` before every `return` on error
- Check `HasError()` immediately after every `Diagnostics.Append()` call
- Include `err.Error()` in the diagnostic detail string

## State Management

**Never:**
- Write partial state (always write all fields or none)
- Read from `req.State` in Create (use `req.Plan`)
- Read from `req.Plan` in Delete or Read (use `req.State`)
- Read from `req.State` in a data source Read (use `req.Config`)
- Change state layout (field names, types, nesting) without a state migration

**Always:**
- Use `resp.State.Set(ctx, &data)` (pointer, not value) and append its diagnostics
- Populate `id` from `Metadata.Ref` of the fully provisioned resource (from polling result)

## Mapping Functions

**Never:**
- Re-implement logic already in `types.go` helpers (`fromTime`, `fromRefPtr`, etc.)
- Read `tenant` from the Terraform model inside `xxxFromModel()` — always pass it as a parameter
- Duplicate mapping logic between resource and data source models without extracting a shared helper

**Always:**
- Return `diag.Diagnostics` from `xxxToXxxModel()` functions, even if currently empty
- Check null/unknown before calling `.ValueString()` on optional attributes
- Cast SDK enum strings explicitly: `sdk.SomeEnum(data.Field.ValueString())`

## Tests

**Never:**
- Use mocking for unit tests in this provider
- Import `terraform-plugin-sdk/v2/helper/acctest` or `terraform-plugin-sdk/v2/helper/resource` (depguard blocks these)
- Write acceptance tests that depend on prior test state (each test must be independently runnable)

**Always:**
- Write unit tests for every new `xxxToXxxModel()` and `xxxFromModel()` function
- Cover null/zero/empty edge cases in unit tests
- Write acceptance tests that verify all computed fields after create

## Package and Module

**Never:**
- Create sub-packages inside `internal/provider/`
- Import `terraform-plugin-sdk/v2` (depguard blocks this)
- Add dependencies without review

## Conventions

**Never:**
- Deviate from the established file naming convention (`resource_xxx.go`, `datasource_xxx.go`)
- Omit the compile-time interface check (`var _ resource.Resource = (*XxxResource)(nil)`)
- Use a different error message format than the established pattern

**Always:**
- Follow the `Configure()` guard pattern: check `req.ProviderData == nil` and type-assert to `clients`
- Register new resources/data sources in `provider.go` `Resources()` / `DataSources()` lists
- Run `make generate` after schema changes to keep docs in sync
