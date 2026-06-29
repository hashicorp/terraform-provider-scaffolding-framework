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

If adding ImportState to an existing resource:
1. Implement resource.ResourceWithImportState on the resource struct
2. Use the resource's `name` as the import ID (verify the API can look up by name)
3. Implement ImportState to call resource.ImportStatePassthroughID() if name = id, or populate state manually
4. Add an acceptance test step with ImportState: true, ImportStateVerify: true
5. Update ai/known-issues.md to remove the relevant known issue entry

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
