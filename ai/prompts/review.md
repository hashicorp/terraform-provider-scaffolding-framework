# Prompt: Code Review

Use this prompt to review a PR or diff for correctness, convention compliance, and safety.

---

## Prompt

```
You are performing a code review for the terraform-provider-seca repository.

Read these documents before reviewing:
- ai/guardrails.md — hard rules; any violation is a blocker
- ai/review-checklist.md — work through every item on this checklist
- ai/provider-conventions.md — naming and schema conventions
- ai/async-operations.md — the mandatory polling pattern
- ai/testing-strategy.md — what tests are required

## Review Scope

[Paste the diff or describe the PR here]

## Review Instructions

1. Work through ai/review-checklist.md item by item. Mark each as:
   - PASS — satisfied
   - FAIL — violated (explain why and what the correct implementation is)
   - N/A — not applicable to this change

2. After the checklist, provide a summary section with:
   - Blockers (must fix before merge)
   - Suggestions (improvements that are not blocking)
   - Praise (good patterns worth noting)

3. For each FAIL, provide a concrete code example of the correct implementation.

## Focus Areas

Pay special attention to:
- Is state populated from the GetXxxUntilState() result (not CreateOrUpdateXxx())?
- Does Read() read from req.State (not req.Plan)?
- Is there a HasError() check after every Diagnostics.Append()?
- Is tenant passed as a parameter to xxxFromModel() rather than read from the model?
- Are unit tests present for all new mapping functions?
- Are null/zero/empty edge cases covered in unit tests?
- Are any terraform-plugin-sdk/v2 packages imported?
```
