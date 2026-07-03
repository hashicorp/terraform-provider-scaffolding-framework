# AI Development Scaffold

This directory is the authoritative reference for AI-assisted development in this repository. Every document is grounded in the actual implementation — nothing here is generic advice.

## Purpose

These documents onboard AI agents (and human contributors) to work safely, consistently, and correctly in this Terraform provider codebase. Read them before writing, reviewing, or refactoring code.

## Document Map

| Document | When to read |
|---|---|
| [requirements.md](requirements.md) | Supported versions, provider goals, compatibility rules |
| [architecture.md](architecture.md) | Package layout, client model, provider lifecycle |
| [provider-conventions.md](provider-conventions.md) | Naming, schema patterns, resource identity |
| [terraform-framework.md](terraform-framework.md) | Framework-specific usage patterns for this provider |
| [async-operations.md](async-operations.md) | How every write operation waits for `Active` state |
| [coding-guidelines.md](coding-guidelines.md) | Go code standards enforced in this repo |
| [testing-strategy.md](testing-strategy.md) | Unit vs acceptance tests, what to cover |
| [guardrails.md](guardrails.md) | Hard rules — read this before any code change |
| [review-checklist.md](review-checklist.md) | Checklist for reviewing PRs |
| [implementation-checklist.md](implementation-checklist.md) | Checklist for implementing new resources/data sources |
| [known-issues.md](known-issues.md) | Technical debt — do not accidentally work around these |
| [roadmap.md](roadmap.md) | Planned resources and their priority |
| [gap-analysis.md](gap-analysis.md) | Full gap analysis, issue drafts, and phased implementation roadmap |
| [glossary.md](glossary.md) | Domain terminology and SDK type mapping |

## Prompts

Ready-to-use prompts for common AI tasks:

| Prompt | Use for |
|---|---|
| [prompts/implementation.md](prompts/implementation.md) | Implementing a new resource or data source |
| [prompts/feature.md](prompts/feature.md) | Adding a feature to an existing resource |
| [prompts/review.md](prompts/review.md) | Code review of a PR or diff |
| [prompts/bugfix.md](prompts/bugfix.md) | Investigating and fixing a bug |
| [prompts/refactor.md](prompts/refactor.md) | Safe refactoring of existing code |

## Reading Order for a New Contributor

1. [requirements.md](requirements.md) — understand what this provider does
2. [architecture.md](architecture.md) — understand the code structure
3. [async-operations.md](async-operations.md) — understand the most critical pattern
4. [provider-conventions.md](provider-conventions.md) — understand naming and schema rules
5. [guardrails.md](guardrails.md) — understand what never to do
6. [roadmap.md](roadmap.md) — understand planned scope
7. [gap-analysis.md](gap-analysis.md) — full gap analysis, issue drafts, phased roadmap
8. [implementation-checklist.md](implementation-checklist.md) — before writing any new resource
