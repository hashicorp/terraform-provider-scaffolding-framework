# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## AI Development Scaffold

The `ai/` directory is the **authoritative reference** for working in this codebase — every doc is grounded in the actual implementation, not generic advice. Start at [`ai/README.md`](ai/README.md) for the full document map and reading order.

**Before any code change, read [`ai/guardrails.md`](ai/guardrails.md).** It contains the hard rules; violating them breaks the provider.

Route to the relevant doc by task — read it *before* you start, not after:

| Your task | Read first |
|---|---|
| Implementing a new resource / data source | [`ai/prompts/implementation.md`](ai/prompts/implementation.md) → [`ai/implementation-checklist.md`](ai/implementation-checklist.md) |
| Adding a feature to an existing resource | [`ai/prompts/feature.md`](ai/prompts/feature.md) |
| Fixing a bug | [`ai/prompts/bugfix.md`](ai/prompts/bugfix.md) + [`ai/known-issues.md`](ai/known-issues.md) |
| Refactoring | [`ai/prompts/refactor.md`](ai/prompts/refactor.md) |
| Reviewing a PR or diff | [`ai/prompts/review.md`](ai/prompts/review.md) → [`ai/review-checklist.md`](ai/review-checklist.md) |
| Anything touching write/Create/Update/Delete | [`ai/async-operations.md`](ai/async-operations.md) (polling is the most critical pattern) |
| Naming, schema, resource identity questions | [`ai/provider-conventions.md`](ai/provider-conventions.md) |
| Framework API usage | [`ai/terraform-framework.md`](ai/terraform-framework.md) |
| Go code standards | [`ai/coding-guidelines.md`](ai/coding-guidelines.md) |
| Writing tests | [`ai/testing-strategy.md`](ai/testing-strategy.md) |
| Domain terms / SDK type mapping | [`ai/glossary.md`](ai/glossary.md) |

**Keep the scaffold in sync:** when you resolve a [`ai/known-issues.md`](ai/known-issues.md) entry or change a documented convention, update the affected `ai/` docs in the same change (the bugfix/feature prompts spell out which files to touch).
