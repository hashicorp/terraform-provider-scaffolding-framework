# Known Issues and Technical Debt

This document captures observed technical debt and known issues. **Do not fix these without a tracked issue and review.** The purpose of documenting them here is to ensure AI agents do not accidentally work around or worsen these problems.

## Critical Gaps

### 1. No `ImportState` Implementation

**Impact:** Users cannot import existing SECA resources into Terraform state (`terraform import`).

**All resources affected:** `seca_workspace`, `seca_image`, `seca_block_storage`

**What's missing:** None of the resource structs implement `resource.ResourceWithImportState`. The framework's `resource.ImportStatePassthroughID()` would be the minimal implementation for name-based imports.

**Risk of workaround:** Do not add `ImportState` unless the SECA API can look up a resource by its `Metadata.Ref` string. Verify the SDK supports `GetXxx(ref)` where ref is the full `<kind>/<name>` string.

---

## Design Gaps

### 2. Duplicated Mapping Logic Between Resource and Data Source

**Impact:** Every resource has both `xxxToResourceModel` and `xxxToDataSourceModel` with nearly identical code. The only difference is that data source models include `state` from status and use `Computed` for maps.

**Example:** `blockStorageToResourceModel` and `blockStorageToDataSourceModel` share ~80% identical mapping code.

**Future improvement:** Extract a shared mapping for common fields (metadata, labels, timestamps), then add resource/data source-specific fields on top.

---

### 3. No `UseStateForUnknown()` on Computed Fields

**Impact:** On every plan, all Computed fields (`tenant`, `region`, `created_at`, etc.) show as `(known after apply)` even when no change is expected. This produces noisy plans and erodes user trust.

**What's missing:** `planmodifier.UseStateForUnknown()` should be added to Computed fields that will not change after initial creation.

---

### 4. Retry Config Is Coarse-Grained

**Impact:** All resources in a provider instance share the same retry config. A slow-provisioning instance and a fast-provisioning workspace cannot have different polling configs.

**Current behavior:** `retry.delay`, `retry.interval`, `retry.max_attempts` are provider-level only.

**Future improvement:** Consider per-resource `timeouts` blocks using `timeouts.New()` from the framework.

---

### 5. No Structured Logging (`tflog`)

**Impact:** No debug output when `TF_LOG=DEBUG` is set. Debugging API interactions requires network tracing.

**What's missing:** `tflog.Debug(ctx, "...")` calls in Create, Read, Update, Delete, and Configure.

---

### 6. Acceptance Test Cluster Is Hard-Coded

**Location:** `internal/acctest/provider_test.go`

**Problem:** The cluster endpoints (`172.18.0.2:30081`) are hard-coded. Running acceptance tests against a different cluster requires editing source code.

**Improvement:** Read endpoints from environment variables (`SECA_REGION_ENDPOINT`, `SECA_AUTH_ENDPOINT`, etc.).

---

### 7. No `CheckDestroy` in Acceptance Tests

**Impact:** Acceptance tests do not verify that resources are actually deleted from the API after `terraform destroy`. The framework's automatic cleanup may succeed at the provider level while leaving orphaned resources on the API.
