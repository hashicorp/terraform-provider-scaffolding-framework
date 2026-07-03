# Known Issues and Technical Debt

This document captures observed technical debt and known issues. **Do not fix these without a tracked issue and review.** The purpose of documenting them here is to ensure AI agents do not accidentally work around or worsen these problems.

## Design Gaps

### 1. Duplicated Mapping Logic Between Resource and Data Source

**Impact:** Every resource has both `xxxToResourceModel` and `xxxToDataSourceModel` with nearly identical code. The only difference is that data source models include `state` from status and use `Computed` for maps.

**Example:** `blockStorageToResourceModel` and `blockStorageToDataSourceModel` share ~80% identical mapping code.

**Future improvement:** Extract a shared mapping for common fields (metadata, labels, timestamps), then add resource/data source-specific fields on top.

