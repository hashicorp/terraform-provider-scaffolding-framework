# Prompt: Implement a New Resource or Data Source

Use this prompt when implementing a new Terraform resource or data source in this provider.

---

## Context to Provide

Before using this prompt, gather:
1. The Terraform resource name (e.g., `seca_network`)
2. The SDK type name (e.g., `sdk.Network`)
3. The SDK service on RegionalClient or GlobalClient (e.g., `NetworkV1`)
4. The available SDK methods (CreateOrUpdate, Get, GetUntilState, Delete)
5. The resource scope: Global, Tenant, or Workspace
6. The SDK Metadata type (e.g., `sdk.RegionalResourceMetadata`)
7. The list of spec fields with their types

---

## Prompt

```
You are implementing a new Terraform resource/data source for the terraform-provider-seca repository.

Read these documents before writing any code:
- ai/architecture.md — understand the package structure and client model
- ai/provider-conventions.md — naming, schema patterns, Configure() pattern
- ai/async-operations.md — the mandatory Create/Update polling pattern
- ai/coding-guidelines.md — formatting, imports, error messages
- ai/guardrails.md — rules you must not break
- ai/implementation-checklist.md — complete every item on this checklist

## Task

Implement [resource/data source]: seca_<name>

## Resource Details

- Terraform name: seca_<name>
- SDK type: sdk.<Name>
- SDK service: [RegionalClient/GlobalClient].<ServiceV1>
- SDK methods:
  - Create/Update: <service>.CreateOrUpdate<Name>()
  - Read: <service>.Get<Name>()
  - Poll: <service>.Get<Name>UntilState()
  - Delete: <service>.Delete<Name>()
- Resource scope: [Global/Tenant/Workspace]
- Metadata type: sdk.<MetadataType>
- Reference type for SDK calls: secapi.[TenantReference/WorkspaceReference]

## Spec Fields

| Field | SDK type | Required/Optional/Computed |
|---|---|---|
| <field_name> | <sdk_type> | <classification> |

## Files to Create

1. internal/provider/resource_<name>.go
2. internal/provider/datasource_<name>.go (if a data source is also needed)
3. internal/provider/resource_<name>_test.go
4. internal/provider/datasource_<name>_test.go (if data source)
5. internal/acctest/<name>_test.go
6. examples/resources/seca_<name>/resource.tf
7. examples/data-sources/seca_<name>/data-source.tf

## Files to Modify

- internal/provider/provider.go — add to Resources() and/or DataSources()

## Constraints

- Follow the existing file structure exactly (see internal/provider/resource_block_storage.go as the primary reference)
- Do not use terraform-plugin-sdk/v2
- Do not skip the polling step in Create() or Update()
- Write state from the GetXxxUntilState() result, never from CreateOrUpdateXxx()
- Apply RequiresReplace() to name and workspace_id (if workspace-scoped)
- Use retry fields from the resource struct, never hard-code values
- Use helpers from types.go — do not re-implement fromTime, fromRefPtr, etc.
- After implementing, verify the implementation-checklist.md is fully satisfied
```
