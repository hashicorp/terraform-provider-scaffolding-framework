# Prompt: Refactor

Use this prompt when refactoring existing code for maintainability, deduplication, or consistency.

---

## Prompt

```
You are refactoring code in the terraform-provider-seca repository.

Read these documents before making any changes:
- ai/guardrails.md — rules you must not break
- ai/architecture.md — understand what should and should not change
- ai/known-issues.md — the items listed here are candidates for improvement, but each requires care
- ai/coding-guidelines.md — conventions to maintain

## Refactor Scope

[Describe what you want to refactor and why]

## Hard Constraints

You must not:
- Change any attribute name, type, required/optional status, or computed status in any schema
- Remove or rename any public type (XxxResource, XxxModel, etc.) — these are used by the framework
- Change the signature of any method required by the framework (Metadata, Schema, Configure, Create, Read, Update, Delete)
- Change the state layout (field names, types, nesting) without a state upgrader
- Introduce new external dependencies
- Create sub-packages inside internal/provider/

## Safe Refactoring Targets

The following refactors are safe if they do not violate the constraints above:
- Extracting common metadata mapping into a shared function
- Extracting the retry config pattern into a shared Configure helper
- Reducing duplication between xxxToResourceModel and xxxToDataSourceModel
- Improving error message consistency
- Adding missing UseStateForUnknown() plan modifiers to Computed fields (non-breaking addition)

## Verification

After refactoring:
1. Run: go build ./...
2. Run: go test -v -cover ./...
3. Run: make lint
4. Confirm all items in ai/review-checklist.md still pass
5. Confirm no items in ai/guardrails.md are violated
```
