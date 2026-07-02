# Prompt: Add a Feature to an Existing Resource

Use this prompt when extending an existing resource or data source with new functionality (new attributes, import support, plan modifiers, validators, etc.).

---

## Prompt

```
You are adding a feature to an existing resource in the terraform-provider-seca repository.

Read these documents before making any changes:
- ai/guardrails.md — hard constraints on schema changes
- ai/terraform-framework.md — framework features and how they're used in this provider
- ai/provider-conventions.md — naming and schema conventions
- ai/known-issues.md — check if the feature you're adding addresses a known issue

## Feature Request

[Describe the feature: what resource, what new behavior, why]

## Schema Change Rules

Adding attributes:
- Adding Computed attributes is always safe (additive, no breaking change)
- Adding Optional attributes requires setting a reasonable zero/empty default so existing configs continue to work
- Adding Required attributes is ALWAYS a breaking change — do not do it without a major version bump

Changing attributes:
- NEVER change the type of an existing attribute
- NEVER change Optional to Required
- NEVER rename an attribute
- Adding RequiresReplace() to an attribute that did not have it is a breaking change

Removing attributes:
- NEVER remove an existing attribute without a deprecation period and state migration

## Adding ImportState

All three existing resources (`seca_workspace`, `seca_image`, `seca_block_storage`) already implement it — follow their pattern:
1. Implement resource.ResourceWithImportState on the resource struct (add the `_ resource.ResourceWithImportState = (*XxxResource)(nil)` interface assertion)
2. The import ID must carry exactly the identity fields `Read()` looks up (tenant always comes from provider config, never the import ID):
   - Tenant-scoped resource (name is enough): pass the `name` — `resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)`
   - Workspace-scoped resource: use a composite `<workspace_id>/<name>` ID, `strings.Cut` on `/`, and `resp.State.SetAttribute()` for each part (see `resource_block_storage.go`)
3. Name the ImportState receiver `r` (not `resource`) so the method body can reach the `resource` package helpers — the usual `resource` receiver name shadows the package
4. Do NOT passthrough to `id`: `Read()` keys off `name`/`workspace_id`, not the `Metadata.Ref` stored in `id`
5. Add an acceptance test step with `ImportState: true`, `ImportStateVerify: true`, and an explicit `ImportStateId` (the default uses `id`, which is wrong here)
6. Update ai/known-issues.md to remove the relevant known issue entry

## Adding UseStateForUnknown()

If adding UseStateForUnknown() to Computed fields:
1. Only add to fields that will NOT change after initial provisioning (tenant, region, created_at)
2. Do NOT add to fields that the API may update (last_modified_at, state)
3. Import both resource/schema/planmodifier and datasource/schema/planmodifier as needed

## Verification

After implementing:
1. Run: go build ./...
2. Run: go test -v -cover ./...
3. Run: make lint
4. Run: make generate (if schema changed)
5. Verify review-checklist.md is satisfied
```
