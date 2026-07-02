# Known Issues and Technical Debt

This document captures observed technical debt and known issues. **Do not fix these without a tracked issue and review.** The purpose of documenting them here is to ensure AI agents do not accidentally work around or worsen these problems.

## Design Gaps

### 1. Duplicated Mapping Logic Between Resource and Data Source

**Impact:** Every resource has both `xxxToResourceModel` and `xxxToDataSourceModel` with nearly identical code. The only difference is that data source models include `state` from status and use `Computed` for maps.

**Example:** `blockStorageToResourceModel` and `blockStorageToDataSourceModel` share ~80% identical mapping code.

**Future improvement:** Extract a shared mapping for common fields (metadata, labels, timestamps), then add resource/data source-specific fields on top.

---

### 2. Retry Config Is Coarse-Grained

**Impact:** All resources in a provider instance share the same retry config. A slow-provisioning instance and a fast-provisioning workspace cannot have different polling configs.

**Current behavior:** `retry.delay`, `retry.interval`, `retry.max_attempts` are provider-level only.

**Future improvement:** Consider per-resource `timeouts` blocks using `timeouts.New()` from the framework.

---

### 3. No Structured Logging (`tflog`)

**Impact:** No debug output when `TF_LOG=DEBUG` is set. Debugging API interactions requires network tracing.

**What's missing:** `tflog.Debug(ctx, "...")` calls in Create, Read, Update, Delete, and Configure.

