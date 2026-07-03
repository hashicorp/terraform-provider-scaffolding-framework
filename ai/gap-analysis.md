# Gap Analysis & Implementation Roadmap

> **Status:** July 2026. Grounded against `go-sdk@v0.4.2`, `examples/`, `ai/`, and the live provider source.
>
> **Scope:** Foundation + Authorization (compute, networking, storage, IAM). Beta SDK extensions
> (loadbalancer, natgateway, kubernetes, objectstorage) are deferred — see Appendix A.
>
> **GitHub issues:** drafts in §11. Reconcile against existing issues #5–#15 (epic #16) before filing.

---

## 1. Executive Summary

**Overall maturity: early foundation / ~25% of intended scope.**

The provider has a clean, consistent architecture and a well-documented AI scaffold (`ai/`), but exposes only 3 of ~13 planned resources and 5 of ~15 planned data sources. Every implemented resource carries the same set of production-blocking cross-cutting gaps.

**Completion estimate (foundation + auth scope):**

| Dimension | Done | Total | % |
|---|---|---|---|
| Resources | 3 | 13 | 23% |
| Data sources | 5 | 15 | 33% |
| Cross-cutting production-readiness (import, drift, delete-polling, logging, timeouts) | — | — | ~10% |

**Architecture quality: good.** One flat `internal/provider` package, strict naming/schema conventions (`ai/provider-conventions.md`), a correct `GlobalClient → RegionalClient` bootstrap, and a documented async-polling pattern. The design scales to the remaining resources with no structural change required.

**Major risks:**

1. **Production-readiness gaps are systemic, not per-resource** — shipping more resources without first fixing import/drift/delete-polling multiplies the debt. *Fix cross-cutting concerns first.*
2. **No drift detection** — `Read` errors instead of removing out-of-band-deleted resources, making the provider unsafe for real state management.
3. **Delete does not confirm completion** → recreate races. The SDK's `WatchXxxUntilDeleted` observers are available but unused.
4. **Hard-coded acceptance-test cluster** (`172.18.0.2:30081`) blocks CI; regressions in new resources will not be caught.
5. **Mapper duplication** (resource vs data-source) will be replicated 10× as new resources land unless refactored early.

**Bottom line:** the hard architectural problems are solved; the remaining work is high-volume but low-novelty wiring, plus a focused hardening pass. The recommended sequence puts hardening (Phase 1) before breadth (Phases 2–5).

---

## 2. Missing Features (grouped by service)

### Cross-cutting (all existing resources)

- `ImportState` (#5) — no `resource.ResourceWithImportState` implementation anywhere.
- 404 → `RemoveResource` drift handling in `Read` (#6) — out-of-band deletes permanently break plans.
- Delete polling via `WatchXxxUntilDeleted` (#7) — the SDK observer exists; the provider ignores it.
- `UseStateForUnknown` on stable computed fields (#10) — every plan shows `(known after apply)` noise.
- `tflog` structured logging (#12) — `TF_LOG=DEBUG` produces no provider output.
- Per-resource `timeouts` blocks (#11) — all resources share one coarse provider-level retry config.
- Shared mapper refactor (#9) — resource/data-source mappers duplicate ~80% of mapping code.
- Env-driven acctest endpoints (#13) — cluster URL is hard-coded in source.
- Update, `ImportStateVerify`, and `CheckDestroy` acctest steps (#14, #15).

### Compute (`ComputeV1`)

- `seca_instance` resource (CRUD + power lifecycle: Start/Stop/Restart + `GetInstanceUntilPowerState`).
- `seca_instance` data source.
- `seca_instance_sku` data source.

### Networking (`NetworkV1`)

- Resources: `seca_network`, `seca_subnet`, `seca_route_table`, `seca_internet_gateway`, `seca_security_group` (with nested inline rules), `seca_public_ip`, `seca_nic`.
- Data sources for each of the above, plus `seca_network_sku`.

### Authorization (`AuthorizationV1`, `GlobalClient`)

- `seca_role` resource + data source.
- `seca_role_assignment` resource + data source.
- Uses `GlobalClient.AuthorizationV1` — mirrors the `seca_region` data source's GlobalClient usage.

### Reconciliation required

- `examples/resources/*/resource.tf` files reference a computed `resource_provider` attribute (provider name + version from `Metadata.Ref`) that none of the 3 current resources expose. This must be resolved before mass doc generation to avoid `docs/` drift.

---

## 3. Missing Resources

| Resource | SDK backing (verified `secapi/*.go`) | Scope | Key spec attributes | Priority |
|---|---|---|---|---|
| `seca_network` | `NetworkV1.CreateOrUpdateNetwork` | Workspace | `sku_id`, `cidr.ipv4`, computed `additional_cidrs` | P1 |
| `seca_internet_gateway` | `NetworkV1.CreateOrUpdateInternetGateway` | Workspace | computed `egress_only` | P1 |
| `seca_route_table` | `NetworkV1.CreateOrUpdateRouteTable` | Workspace | `network_id`, `routes[]{destination_cidr_block, target_id}` | P1 |
| `seca_subnet` | `NetworkV1.CreateOrUpdateSubnet` | Workspace | `network_id`, `cidr.ipv4`, `route_table_id`, computed `zone`, `sku_id` | P1 |
| `seca_security_group` | `NetworkV1.CreateOrUpdateSecurityGroup` + `…Rule` | Workspace | `rules[]{direction, protocol, ports{list\|from\|to}, source_refs}`, computed `rule_refs` | P2 |
| `seca_public_ip` | `NetworkV1.CreateOrUpdatePublicIp` | Workspace | `version` (IPv4/IPv6), computed `address`, `attached_to` | P2 |
| `seca_nic` | `NetworkV1.CreateOrUpdateNic` | Workspace | `subnet_id`, `addresses[]`, `public_ip_id(s)`, computed `mac_address`, `security_group_ids` | P2 |
| `seca_instance` | `ComputeV1.CreateOrUpdateInstance` + Start/Stop/Restart | Workspace | `sku_id`, `primary_nic_id`, `zone`, `ssh_keys[]`, `boot_volume{device_id}`, `data_volumes`, computed `power_state`, `power_state_since` | P3 |
| `seca_role` | `AuthorizationV1.CreateOrUpdateRole` | Tenant (Global) | `permissions[]{provider, resources[], verb[]}` | P4 |
| `seca_role_assignment` | `AuthorizationV1.CreateOrUpdateRoleAssignment` | Tenant (Global) | `subs[]`, `scopes[]{tenants, regions, workspaces}`, `roles[]` | P4 |

**Note on `security_group_rule`:** The SDK exposes `CreateOrUpdateSecurityGroupRule` as a distinct sub-resource, but `examples/` models rules **inline** inside `seca_security_group`. Recommendation: keep rules inline (matching examples), with rule CRUD managed inside the security-group resource's Create/Update. Do not add a separate `seca_security_group_rule` resource unless the API semantics require independent lifecycle.

---

## 4. Missing Data Sources

| Data source | SDK backing | Priority |
|---|---|---|
| `seca_network_sku` | `NetworkV1.ListSkus` / `GetSku` | P1 |
| `seca_network` | `NetworkV1.GetNetwork` | P1 |
| `seca_internet_gateway` | `NetworkV1.GetInternetGateway` | P1 |
| `seca_route_table` | `NetworkV1.GetRouteTable` | P1 |
| `seca_subnet` | `NetworkV1.GetSubnet` | P1 |
| `seca_security_group` | `NetworkV1.GetSecurityGroup` | P2 |
| `seca_public_ip` | `NetworkV1.GetPublicIp` | P2 |
| `seca_nic` | `NetworkV1.GetNic` | P2 |
| `seca_instance` | `ComputeV1.GetInstance` | P3 |
| `seca_instance_sku` | `ComputeV1.ListSkus` / `GetSku` | P3 |
| `seca_role` | `AuthorizationV1.GetRole` | P4 |
| `seca_role_assignment` | `AuthorizationV1.GetRoleAssignment` | P4 |

---

## 5. Missing Acceptance Tests

| Area | Gap | Tracking |
|---|---|---|
| All existing resources | No Update step (mutate mutable fields, e.g. `size_gb`, `labels`) | #14 |
| All existing resources | No `CheckDestroy` verifying API-side deletion | #15 |
| All existing resources | No `ImportState` / `ImportStateVerify` step | blocked by #5 |
| Acctest harness | Endpoints hard-coded to `172.18.0.2:30081`; cannot run in CI | #13 |
| All new resources | Full CRUD + import + update + `CheckDestroy` per resource | GA-20, GA-27, GA-32, GA-37 |
| `seca_instance` | Power-state lifecycle (start / stop / restart) | GA-32 |
| Negative cases | Invalid SKU / invalid CIDR / missing workspace — assert error messages | GA-41 |

---

## 6. Missing Unit Tests

| Area | Gap |
|---|---|
| Existing mappers | Only happy-path `xxxToResourceModel`; no null/unknown-field, empty-map, nil-pointer-ref cases |
| 404 handling | No test that `Read` calls `RemoveResource` on not-found (blocked by GA-1/#6) |
| Provider `Configure` | No test for retry-default fallback, nil `authorization_v1`, wrong provider-data type |
| `types.go` helpers | `numberToDuration` / `numberToInt` edge cases partially covered (`types_test.go`); add fractional, zero, null |
| New resources | `fooFromModel` / `fooToModel` round-trip tests, nested-object mapping (routes, rules, scopes) |
| Schema validators | Once CIDR / enum / port-range validators are added, unit-test each |

---

## 7. Documentation Gaps

| Gap | Detail |
|---|---|
| Registry docs for 10 resources / 12 data sources | `docs/` only covers workspace, image, block_storage (+ 5 data sources). `make generate` must run after every new resource. |
| No import guide | No `terraform import` examples; add `docs/guides/import.md` once GA-3 / #5 lands. |
| No authentication guide | Token acquisition, `global_providers` endpoints, tenant/region selection are undocumented as a user-facing guide. |
| No troubleshooting guide | Async timeout tuning (`retry`), eventual-consistency behavior, common API errors. |
| `examples/` ↔ schema drift | `resource_provider` output in examples has no schema backing; regenerate examples per resource (GA-10). |
| Provider index thin | `docs/index.md` should document the `retry` block, non-goals (single-region per block), and the `id` reference format. |

---

## 8. Technical Debt

### Critical

| ID | Description | Fix |
|---|---|---|
| TD-C1 (#6) | **No 404 handling in `Read`.** Out-of-band deletes cause permanent plan errors. | Detect not-found via `secapi/errors.go`; call `resp.State.RemoveResource(ctx)` and return. |
| TD-C2 (#7) | **Delete does not confirm completion.** Recreate races possible. | After `DeleteXxx`, call `WatchXxxUntilDeleted` with `ResourceObserverConfig`. Note: no `Deleted` state exists — deletion is confirmed by the Watch observer (404), not a state poll. |
| TD-C3 (#5) | **No `ImportState`.** Users cannot `terraform import` existing resources. | Implement `resource.ResourceWithImportState`; passthrough on `id` once confirmed `GetXxx` accepts a parsed `<kind>/<name>` ref. |

### High

| ID | Description | Fix |
|---|---|---|
| TD-H1 (#8) | Copy-paste error: `resource_workspace.go:168` says `"Error updating workspace"` in `Create()`. | One-line string fix. |
| TD-H2 (#13) | Hard-coded acctest cluster (`172.18.0.2:30081`) blocks CI. | Read from `SECA_REGION_ENDPOINT`, `SECA_AUTH_ENDPOINT` env vars. |
| TD-H3 | `examples/` reference `resource_provider` attribute with no schema backing. | Decide add-vs-remove; apply uniformly before doc generation (GA-10). |
| TD-H4 (#9) | Mapper duplication (~80% shared code between resource and data-source mappers). | Extract shared metadata/labels/timestamps helper; will be replicated 10× if not refactored before breadth work. |

### Medium

| ID | Description |
|---|---|
| TD-M1 (#10) | No `UseStateForUnknown` on stable computed fields → noisy plans on every apply. |
| TD-M2 (#12) | No `tflog` structured logging → `TF_LOG=DEBUG` produces no provider-level output. |
| TD-M3 (#11) | Provider-level retry only; no per-resource `timeouts` blocks (instances provision far slower than workspaces). |
| TD-M4 | No schema validators (CIDR, IP-version enum, port range 1–65535). |
| TD-M5 (#14/#15) | Acceptance tests lack Update steps and `CheckDestroy`. |

### Low

| ID | Description |
|---|---|
| TD-L1 | Inconsistent receiver names (`resource` vs `r`) across resource files. |
| TD-L2 | Error detail strings built by hand per method — candidate for a small `diagError` helper. |
| TD-L3 | `Configure` error message says `"Expected sdk.Clients"` but the actual type name is `clients`. |

---

## 9. Provider Improvement Suggestions

> These are architectural recommendations. No code samples here — see `ai/implementation-checklist.md` for implementation patterns.

1. **Shared metadata mapper.** Extract common metadata / labels / timestamps mapping into a helper reused by both resource and data-source mappers (addresses TD-H4 / #9). Apply the `datasource_region.go:143–172` nested-object pattern (`attr-type map` + `types.ListValueFrom`) for all list-of-objects fields (`routes`, `rules`, `scopes`, `permissions`).

2. **Async lifecycle "kit".** A thin internal helper for the repeated async lifecycle — submit mutation → `GetXxxUntilState` → map → set state; delete → `WatchXxxUntilDeleted` — so each new resource's CRUD body becomes mapping-only. The SDK already provides the observers.

3. **Framework `timeouts` blocks.** Adopt `timeouts.New()` per resource (create / update / delete) mapped onto the observer config, replacing/augmenting the coarse provider-level retry (TD-M3 / #11). Instances will need 5–10× longer timeouts than workspaces.

4. **Centralize not-found detection.** One `isNotFound(err) bool` helper over `secapi/errors.go`, shared by every `Read` and by the delete-watch path (TD-C1 / #6).

5. **Reusable validators.** CIDR-block, IP-version enum (`IPv4`/`IPv6`), port-range (1–65535), direction enum (`ingress`/`egress`). Document them in `ai/provider-conventions.md` so every networking resource applies them uniformly.

6. **`tflog` convention.** Standard debug lines in `Configure`, `Create`, `Read`, `Update`, `Delete` (request ref, poll attempt count, final state). Cheap, high debugging payoff (TD-M2 / #12).

7. **Reconcile `resource_provider`.** Surface the SDK's provider/version segment as a computed attribute on all resources *or* remove it from `examples/` — decide once in GA-10 before `make generate` bakes drift into `docs/`.

---

## 10. Prioritized Roadmap

### Phase 1 — Production hardening (cross-cutting)

**Rationale:** Every new resource inherits these patterns. Fixing them first prevents 10× duplication of debt; codifying them as convention protects all subsequent work. Also unblocks CI (env-driven acctests).

**Scope:** not-found helper + `Read` drift (#6), delete-watch (#7), `ImportState` (#5), workspace diagnostic copy-paste fix (#8), env-driven acctests (#13), Update + `CheckDestroy` + `ImportStateVerify` acctest steps (#14/#15), shared mapper refactor (#9), `UseStateForUnknown` (#10), `tflog` (#12), `resource_provider` reconciliation (GA-10).

**Dependencies:** none (all SDK primitives exist). **Effort:** M–L. **Risk:** low-med (must confirm SDK not-found signal in `secapi/errors.go` and verify import ref parsing before GA-3).

---

### Phase 2 — Networking core

**Resources + data sources:** `seca_network`, `seca_internet_gateway`, `seca_route_table`, `seca_subnet` + matching data sources + `seca_network_sku`.

**Rationale:** Highest immediate user value; prerequisite for NICs and instances. Internal ordering: `network` must exist before `route_table` and `subnet`.

**Key challenge:** `seca_route_table` has a nested `routes[]` list-of-objects — use the region data-source's `types.ListValueFrom` pattern. `seca_network` needs a `cidr` nested object (`ipv4` string). CIDR validators apply here first.

**Dependencies:** Phase 1 conventions. **Effort:** L. **Risk:** medium.

---

### Phase 3 — Networking edge

**Resources + data sources:** `seca_security_group` (inline rules), `seca_public_ip`, `seca_nic` + matching data sources.

**Rationale:** Required substrate for compute instances (NIC + public IP + security group must exist before an instance can be created).

**Key challenge:** `seca_security_group` rules have a polymorphic `ports` object (`list`, `from`, or `from+to`). This requires a nested `SingleNestedAttribute` or custom object type with careful null/unknown handling. Port-range validators apply here.

**Dependencies:** Phase 2 (subnet for NIC). **Effort:** L. **Risk:** medium-high (SG rule polymorphism).

---

### Phase 4 — Compute

**Resources + data sources:** `seca_instance` (+ power lifecycle), `seca_instance` data source, `seca_instance_sku` data source.

**Rationale:** The capstone that ties storage + networking together and makes `examples/use-cases/usage.tf` fully applyable end-to-end.

**Key challenge:** Instance power-state modeling. Options: (a) computed `power_state` only (read-only), or (b) desired-state attribute that triggers Start/Stop during Update. Document the decision in GA-29 before implementing. Boot-volume and data-volumes are nested objects requiring the established pattern.

**Dependencies:** Phases 2–3 + existing storage resources. **Effort:** L. **Risk:** high (power lifecycle semantics, `GetInstanceUntilPowerState`).

---

### Phase 5 — Authorization / IAM

**Resources + data sources:** `seca_role`, `seca_role_assignment` + data sources.

**Rationale:** Independent of compute/network — can be parallelized with Phases 2–4 by a second contributor. Uses `GlobalClient.AuthorizationV1` (mirrors `datasource_region.go`).

**Key challenge:** `seca_role.permissions` and `seca_role_assignment.scopes` are nested lists-of-objects. Use the established nested-mapping pattern. Note that authorization resources are **tenant-scoped via GlobalClient**, not workspace-scoped via RegionalClient.

**Dependencies:** Phase 1 only. **Effort:** M. **Risk:** medium.

---

### Phase 6 — Docs & polish

**Scope:** Regenerate all `docs/` (`make generate`), author import/authentication/troubleshooting guides, add reusable schema validators, write negative acceptance tests, implement per-resource `timeouts` blocks, GA checklist.

**Dependencies:** all prior phases. **Effort:** M. **Risk:** low.

---

## 11. GitHub Issues (draft)

> Existing issues #5–#15 and epic #16 cover Phase-1 cross-cutting items. The entries below **reference** those issues and add implementation detail, then list **new** issues for Phases 2–6.
>
> **Issue template fields (apply to every issue):** Title · Description · Background · Requirements · Implementation Notes · Acceptance Criteria · Testing Requirements · Documentation Requirements · Dependencies · Labels · Estimated Complexity · Suggested Milestone · Definition of Done.

---

### Phase 1 — Hardening (refines #5–#15)

**GA-1 (→#6) — `Read` drift: not-found → `RemoveResource`**
- Background: `GetXxx` returns an error when a resource is deleted out-of-band; `Read` converts it to a diagnostic, leaving state permanently broken.
- Requirements: inspect `secapi/errors.go` to identify the not-found error type/code; add `isNotFound(err) bool` in `clients.go`; update `Read` on `seca_workspace`, `seca_image`, `seca_block_storage`.
- Acceptance criteria: delete a resource via API; `terraform plan` proposes recreate (not error).
- Labels: `tech-debt`, `critical` · Complexity: S · Milestone: M1

**GA-2 (→#7) — Delete polling via `WatchXxxUntilDeleted`**
- Background: `Delete` returns as soon as `DeleteXxx` API call succeeds; the resource may still be terminating. A same-named recreate may conflict.
- Requirements: after `DeleteXxx`, call `WatchXxxUntilDeleted(ctx, ref, ResourceObserverConfig{...retry fields...})`; handle timeout error as diagnostic. Note: no `Deleted` `ResourceState` exists — deletion is confirmed by the Watch observer (404).
- Acceptance criteria: rapid destroy+create of same-named resource succeeds without conflict.
- Labels: `tech-debt`, `critical` · Complexity: S · Milestone: M1

**GA-3 (→#5) — `ImportState` on all 3 existing resources**
- Background: no `resource.ResourceWithImportState` implementation exists.
- Requirements: confirm `GetXxx` can resolve a `<kind>/<name>` ref (e.g. `block-storages/my-vol`) by inspecting `secapi/references.go`; implement `ImportState` using `resource.ImportStatePassthroughID`; update `Read` to accept both name and full ref as identity.
- Acceptance criteria: `terraform import seca_block_storage.x block-storages/my-vol` produces valid state.
- Labels: `feature`, `critical` · Complexity: M · Milestone: M1

**GA-4 (→#8) — Fix diagnostic message in `resource_workspace.go:168`**
- Background: `Create()` emits `"Error updating workspace"` instead of `"Error creating workspace"`.
- Labels: `bug`, `good-first-issue` · Complexity: XS · Milestone: M1

**GA-5 (→#13) — Env-driven acctest endpoints**
- Background: `internal/acctest/provider_test.go` hard-codes `172.18.0.2:30081`; CI cannot run against a different cluster.
- Requirements: read `SECA_REGION_ENDPOINT`, `SECA_AUTH_ENDPOINT`, `SECA_TOKEN`, `SECA_TENANT`, `SECA_REGION` from env vars; skip acctest suite if required vars are unset.
- Labels: `testing`, `ci` · Complexity: S · Milestone: M1

**GA-6 (→#14/#15) — Expand existing acctests (update + import + CheckDestroy)**
- Background: existing acctests only cover create + read; no update, import, or destroy verification.
- Requirements: for each of `seca_workspace`, `seca_image`, `seca_block_storage`, add: (a) a second `TestStep` that updates a mutable field (e.g. `labels`); (b) `resource.TestCheckResourceAttr` assertions; (c) `ImportStateVerify`; (d) `CheckDestroy` function.
- Labels: `testing` · Complexity: M · Milestone: M1

**GA-7 (→#9) — Extract shared metadata mapper**
- Background: `blockStorageToResourceModel` and `blockStorageToDataSourceModel` share ~80% identical mapping code; same pattern will repeat for every new resource.
- Requirements: add `mapMetadata(m *sdk.RegionalWorkspaceResourceMetadata, model *…)` and similar helpers; update existing 3 resources; document convention in `ai/provider-conventions.md`.
- Labels: `refactor` · Complexity: M · Milestone: M1

**GA-8 (→#10) — `UseStateForUnknown` on stable computed fields**
- Background: `tenant`, `region`, `created_at`, `id` show as `(known after apply)` on every plan even when unchanged, producing plan noise.
- Requirements: add `stringplanmodifier.UseStateForUnknown()` to these fields on all resources.
- Labels: `enhancement`, `ux` · Complexity: S · Milestone: M1

**GA-9 (→#12) — `tflog` structured debug logging**
- Background: `TF_LOG=DEBUG` produces no provider-level output; debugging requires network tracing.
- Requirements: add `tflog.Debug` calls in `Configure`, `Create`, `Read`, `Update`, `Delete` across provider; log resource ref, operation, and poll attempt count.
- Labels: `enhancement`, `observability` · Complexity: S · Milestone: M1

**GA-10 (new) — Reconcile `resource_provider` attribute**
- Background: `examples/resources/*/resource.tf` emit `resource_provider` as a computed output, but no resource schema includes this attribute.
- Requirements: (a) inspect SDK `Metadata.Ref` to determine if provider/version is extractable; (b) decide add-vs-remove; (c) if adding, implement as `Computed: true` string on all resources; if removing, update examples. Apply decision before `make generate`.
- Labels: `bug`, `docs` · Complexity: S · Milestone: M1

---

### Phase 2 — Networking core (new)

**GA-11 — `seca_network` resource**
- Requirements: CRUD + async poll (`GetNetworkUntilState`); attributes `sku_id` (RequiresReplace), `cidr` object (`ipv4` string, RequiresReplace), computed `additional_cidrs`.
- Implementation notes: use `NetworkV1.CreateOrUpdateNetwork`; model `cidr` as `SingleNestedAttribute`; add CIDR string validator.
- Labels: `feature`, `networking` · Complexity: M · Milestone: M2

**GA-12 — `seca_network` data source**
- Requirements: read by `(workspace_id, name)`; all attributes Computed.
- Labels: `feature`, `networking` · Complexity: S · Milestone: M2

**GA-13 — `seca_network_sku` data source**
- Requirements: mirror `datasource_storage_sku.go`; use `NetworkV1.GetSku(ctx, TenantReference{name})`.
- Labels: `feature`, `networking` · Complexity: S · Milestone: M2

**GA-14 — `seca_internet_gateway` resource**
- Requirements: CRUD + poll; no spec attributes; computed `egress_only`.
- Labels: `feature`, `networking` · Complexity: S · Milestone: M2

**GA-15 — `seca_internet_gateway` data source**
- Labels: `feature`, `networking` · Complexity: XS · Milestone: M2

**GA-16 — `seca_route_table` resource**
- Requirements: `network_id` (RequiresReplace); `routes` as `ListNestedAttribute` of `{destination_cidr_block, target_id}`; CRUD + poll.
- Implementation notes: follow `datasource_region.go:79–96` nested pattern; add CIDR validator on `destination_cidr_block`.
- Labels: `feature`, `networking` · Complexity: M · Milestone: M2

**GA-17 — `seca_route_table` data source**
- Labels: `feature`, `networking` · Complexity: S · Milestone: M2

**GA-18 — `seca_subnet` resource**
- Requirements: `network_id` (RequiresReplace), `cidr` object, `route_table_id` (Optional, mutable); computed `zone`, `sku_id`.
- Labels: `feature`, `networking` · Complexity: M · Milestone: M2

**GA-19 — `seca_subnet` data source**
- Labels: `feature`, `networking` · Complexity: S · Milestone: M2

**GA-20 — Phase-2 networking core acceptance tests**
- Requirements: one acctest file per resource; CRUD + update + import + `CheckDestroy`; acctest resources must be created in dependency order (network → gateway + route_table → subnet).
- Labels: `testing`, `networking` · Complexity: L · Milestone: M2

---

### Phase 3 — Networking edge (new)

**GA-21 — `seca_security_group` resource**
- Requirements: `rules` as `ListNestedAttribute` of `{direction, protocol, ports{list[]|from|to}, source_refs[]}`; inline rule CRUD via `CreateOrUpdateSecurityGroupRule`; computed `rule_refs`.
- Implementation notes: `ports` is a polymorphic object — model as `SingleNestedAttribute` with all three sub-fields Optional; add `direction` enum validator (`ingress`/`egress`), port-range validator.
- Labels: `feature`, `networking` · Complexity: L · Milestone: M2

**GA-22 — `seca_security_group` data source**
- Labels: `feature`, `networking` · Complexity: S · Milestone: M2

**GA-23 — `seca_public_ip` resource**
- Requirements: `version` (enum `IPv4`/`IPv6`, RequiresReplace); computed `address`, `attached_to`.
- Labels: `feature`, `networking` · Complexity: S · Milestone: M2

**GA-24 — `seca_public_ip` data source**
- Labels: `feature`, `networking` · Complexity: XS · Milestone: M2

**GA-25 — `seca_nic` resource**
- Requirements: `subnet_id` (RequiresReplace), `addresses` (list of strings), `public_ip_id` (Optional); computed `public_ip_ids`, `security_group_ids`, `mac_address`, `sku_id`.
- Labels: `feature`, `networking` · Complexity: M · Milestone: M2

**GA-26 — `seca_nic` data source**
- Labels: `feature`, `networking` · Complexity: S · Milestone: M2

**GA-27 — Phase-3 networking edge acceptance tests**
- Requirements: SG + public_ip + NIC acctests; NIC test depends on Phase-2 subnet.
- Labels: `testing`, `networking` · Complexity: L · Milestone: M2

---

### Phase 4 — Compute (new)

**GA-28 — `seca_instance` resource (CRUD)**
- Requirements: `sku_id` (RequiresReplace), `primary_nic_id` (RequiresReplace), `zone` (RequiresReplace), `ssh_keys` (list), `boot_volume` object (`device_id`), `data_volumes` list (Optional); CRUD + `GetInstanceUntilState`.
- Labels: `feature`, `compute` · Complexity: L · Milestone: M3

**GA-29 — Instance power-state lifecycle**
- Background: SDK provides `StartInstance`, `StopInstance`, `RestartInstance`, `GetInstanceUntilPowerState`.
- Requirements: decide whether to expose desired power state as a Terraform attribute (Optional, triggers Start/Stop in Update) or as a separate resource. Document design decision in this issue before implementing. Computed `power_state`, `power_state_since` regardless.
- Labels: `feature`, `compute`, `design-decision` · Complexity: L · Milestone: M3

**GA-30 — `seca_instance` data source**
- Labels: `feature`, `compute` · Complexity: S · Milestone: M3

**GA-31 — `seca_instance_sku` data source**
- Requirements: mirror `datasource_storage_sku.go`; use `ComputeV1.GetSku`.
- Labels: `feature`, `compute` · Complexity: S · Milestone: M3

**GA-32 — Compute acceptance tests**
- Requirements: full instance lifecycle test (create → read → power ops → destroy); requires NIC, block storage, image as prerequisites; `CheckDestroy`.
- Labels: `testing`, `compute` · Complexity: L · Milestone: M3

---

### Phase 5 — Authorization (new)

**GA-33 — `seca_role` resource**
- Requirements: use `GlobalClient.AuthorizationV1`; `permissions` as `ListNestedAttribute` of `{provider, resources[], verb[]}`; CRUD + `GetRoleUntilState`; tenant-scoped (no `workspace_id`).
- Implementation notes: mimic `datasource_region.go`'s `GlobalClient` usage; `Configure` assigns `clients.GlobalClient`.
- Labels: `feature`, `iam` · Complexity: M · Milestone: M4

**GA-34 — `seca_role` data source**
- Labels: `feature`, `iam` · Complexity: S · Milestone: M4

**GA-35 — `seca_role_assignment` resource**
- Requirements: `subs` (list), `scopes` nested list (`{tenants, regions, workspaces}`), `roles` (list); `GetRoleAssignmentUntilState`.
- Labels: `feature`, `iam` · Complexity: M · Milestone: M4

**GA-36 — `seca_role_assignment` data source**
- Labels: `feature`, `iam` · Complexity: S · Milestone: M4

**GA-37 — Authorization acceptance tests**
- Notes: uses `GlobalClient`; ensure acctest provider factory supplies a token with authorization-admin permissions.
- Labels: `testing`, `iam` · Complexity: M · Milestone: M4

---

### Phase 6 — Docs & polish (new)

**GA-38 — Regenerate `docs/` for all new resources/data sources**
- Requirements: run `make generate` after each Phase-2–5 issue lands; CI fails if `docs/` diverges.
- Labels: `docs` · Complexity: XS per resource · Milestone: M5

**GA-39 — User guides: import, authentication, troubleshooting**
- Requirements: `docs/guides/import.md` (after GA-3); `docs/guides/authentication.md`; `docs/guides/troubleshooting.md` (retry tuning, eventual-consistency, common errors).
- Labels: `docs` · Complexity: M · Milestone: M5

**GA-40 — Reusable schema validators**
- Requirements: CIDR-block, IP-version enum, port-range (1–65535), direction enum; unit tests for each; apply retroactively to networking resources; document in `ai/provider-conventions.md`.
- Labels: `enhancement`, `quality` · Complexity: M · Milestone: M5

**GA-41 — Negative acceptance tests**
- Requirements: invalid SKU → expect error; invalid CIDR → expect error; missing workspace → expect error. Cover at least `seca_network`, `seca_block_storage`, `seca_instance`.
- Labels: `testing` · Complexity: M · Milestone: M5

**GA-42 — Per-resource `timeouts` blocks**
- Requirements: add `timeouts.New()` (create/update/delete) to all resources; map onto `ResourceObserverUntilValueConfig`; document defaults in `docs/` and `ai/async-operations.md`.
- Labels: `enhancement` · Complexity: M · Milestone: M5

---

## 12. Milestones

| Milestone | Contents | Issues |
|---|---|---|
| **M1 Provider Foundation** | Cross-cutting hardening on existing 3 resources | GA-1…GA-10 (refines #5–#15) |
| **M2 Networking** | network / igw / route_table / subnet / SG / public_ip / NIC + data sources + skus | GA-11…GA-27 |
| **M3 Compute** | instance (+ power lifecycle) + data sources + instance_sku | GA-28…GA-32 |
| **M4 Identity** | role, role_assignment + data sources | GA-33…GA-37 |
| **M5 Documentation** | docs regen, guides, validators, negative tests, timeouts | GA-38…GA-42 |
| **M6 GA Release** | full acctest matrix green in CI, `docs/` parity, version bump, registry publish | release checklist |

---

## 13. Recommended Development Order

1. **M1 — Harden existing 3 resources** (GA-1…GA-10). Establishes the reusable conventions before any breadth work, unblocks CI via env-driven acctests, and ensures all new resources are born production-ready.

2. **M2 — Networking core, dependency order** (network → internet_gateway + route_table → subnet, then edge: public_ip, nic, security_group). Prerequisite for compute; highest immediate user value.

3. **M3 — Compute** (instance + power lifecycle). Capstone that makes `examples/use-cases/usage.tf` fully apply end-to-end.

4. **M4 — Identity** (role, role_assignment). Independent of compute/network; can be parallelized by a second contributor since it depends only on M1.

5. **M5/M6 — Docs & polish → GA.** Continuous per resource as each Phase lands; finalize as a dedicated milestone sweep.

**Ordering rationale:** dependency correctness (network → subnet → nic → instance), Terraform-workflow safety (import + drift detection before breadth), CI coverage (acctests unblocked at M1), and API maturity (stable v1 before any beta extension).

---

## Appendix A — Out of scope (SDK `v1beta1` extensions)

`go-sdk@v0.4.2` ships beta clients with no `examples/` backing: **loadbalancer**, **natgateway**, **kubernetes**, **objectstorage**, `wellknown`, `activitylog`. These are deferred until they stabilize and gain example HCL. Do not implement speculatively.

## Appendix B — Assumptions

- The `<kind>/<name>` `id` ref (e.g. `block-storages/my-vol`) is sufficient for `GetXxx`-based import (must be confirmed against `secapi/references.go` during GA-3).
- `seca_security_group` rules are modeled **inline** (matching `examples/`), not as a standalone `seca_security_group_rule` resource, despite the SDK exposing separate `SecurityGroupRule` CRUD.
- `resource_provider` reconciliation (GA-10) is expected to resolve as **add** (the SDK `Metadata.Ref` contains the provider/version prefix); to be confirmed during GA-10.
