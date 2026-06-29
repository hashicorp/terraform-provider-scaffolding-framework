# Prompt: Bug Fix

Use this prompt when investigating and fixing a reported bug.

---

## Prompt

```
You are fixing a bug in the terraform-provider-seca repository.

Read these documents before making any changes:
- ai/architecture.md — understand the code structure
- ai/async-operations.md — common source of bugs
- ai/guardrails.md — rules you must not break while fixing
- ai/known-issues.md — check if this bug is a known issue before attempting a fix

## Bug Report

[Describe the bug, error messages, and reproduction steps]

## Investigation Steps

1. Identify the affected resource or data source
2. Determine which CRUD method is involved (Create/Read/Update/Delete)
3. Check the mapping functions (xxxToModel, xxxFromModel) for the affected fields
4. Check if the bug is related to async polling (state written from wrong source)
5. Check if the bug is related to null/unknown handling
6. Check if the bug is a known issue in ai/known-issues.md

## Fix Constraints

- Do not change schema attribute names, types, or required/optional status unless the bug IS the schema being wrong
- Do not change state layout (field ordering, nesting) without adding a state migration
- Do not skip the polling step as a "quick fix" for async issues
- Do not add a workaround for a known issue — fix it properly or leave it and document it
- After fixing, verify the fix does not introduce any item in ai/guardrails.md

## Testing the Fix

- Add or update the relevant unit test in internal/provider/ to cover the bug case
- If the bug is reproducing in acceptance tests, add a test step that would have caught it
- Run: go test -v -run TestXxx ./internal/provider/
```
