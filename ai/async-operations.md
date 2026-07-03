# Async Operations

## Why Resources Are Async

The SECA API is eventually consistent. `CreateOrUpdateXxx()` returns as soon as the API accepts the request, but the resource may still be provisioning. The provider must poll until the resource reaches `Active` state before writing state to Terraform. Without this, Terraform would read back an incomplete resource on the next plan.

## The Polling Pattern

Every `Create()` and `Update()` in this provider follows these exact steps:

```go
// Step 1: Submit the mutation
result, err := resource.client.StorageV1.CreateOrUpdateXxx(ctx, sdkObject)
if err != nil {
    resp.Diagnostics.AddError("Error creating Xxx", "...\nError: "+err.Error())
    return
}

// Step 2: Build the reference for polling
ref := secapi.TenantReference{          // or WorkspaceReference for workspace-scoped
    Tenant: secapi.TenantID(result.Metadata.Tenant),
    Name:   result.Metadata.Name,
}

// Step 3: Configure polling
// resource.retry holds the provider-level values; .with(data.Retry) overlays
// any per-resource `retry` block; .untilState builds the observer config.
config := resource.retry.with(data.Retry).untilState(sdk.ResourceStateActive)

// Step 4: Poll until active (or fail)
result, err = resource.client.StorageV1.GetXxxUntilState(ctx, ref, config)
if err != nil {
    resp.Diagnostics.AddError("Error reading Xxx", "...\nError: "+err.Error())
    return
}
```

**The result from Step 4 (not Step 1) must be used to populate Terraform state.** The state must reflect the fully provisioned resource, not the intermediate response from the create call.

## Retry Configuration

Retry parameters resolve through three layers, each overriding the previous **per field**:

1. **Hardcoded defaults** in `clients.go` (`defaultRetryDelay` etc.).
2. **Provider `retry` block** — replaces any field it sets for all resources; passed to each resource via the `clients` struct.
3. **Per-resource `retry` block** — replaces any field it sets for that one resource only.

| Parameter | Default | Attribute (provider & resource) |
|---|---|---|
| `Delay` | `30s` | `retry.delay` (seconds) |
| `Interval` | `10s` | `retry.interval` (seconds) |
| `MaxAttempts` | `5` | `retry.max_attempts` |

`Delay` is the initial wait before the first poll (allows the API time to begin provisioning). `Interval` is the wait between subsequent polls.

Each resource stores the resolved provider-level values in `resource.retry` (a `retryConfig`, set in `Configure()`). The optional per-resource block lives on the model as `Retry *SecaRetryModel` and is declared in the schema with the shared `retryResourceSchema()` helper. In Create/Update/Delete, `resource.retry.with(data.Retry)` overlays the per-resource block on top of the inherited values before building the observer config (`.untilState(...)` for Create/Update, `.observer()` for Delete). All of this lives in `retry.go`.

```hcl
provider "seca" {
  retry = {
    interval = 15            # applies to every resource
  }
}

resource "seca_block_storage" "slow" {
  # ...
  retry = {
    max_attempts = 40        # overrides ONLY max_attempts for this resource;
                             # delay & interval still come from the provider/defaults
  }
}
```

## Reference Type Selection

The polling reference type depends on the resource scope:

| Resource scope | Reference type | Fields |
|---|---|---|
| Tenant | `secapi.TenantReference` | `Tenant`, `Name` |
| Workspace | `secapi.WorkspaceReference` | `Tenant`, `Workspace`, `Name` |

Always use the **result from the initial API call** to populate the reference (e.g., `result.Metadata.Tenant`), not the values from the Terraform model. This is because the API may assign or normalize field values.

## Delete Operations

`Delete()` submits the deletion with `DeleteXxx()`, then polls `WatchXxxUntilDeleted(ctx, ref, config)` — using the same reference and the same resolved retry config (`resource.retry.with(data.Retry).observer()`) as the Create/Update polling — before returning. This ensures the resource is fully gone on the API side, so a subsequent create of a same-named resource does not conflict. On a polling error, surface it with the read verb: `resp.Diagnostics.AddError("Error reading Xxx", "...while waiting for the Xxx to become deleted.\nError: "+err.Error())`.

## What NOT to Do

- Never write state from the Step 1 result. Always use the Step 4 (poll) result.
- Never skip the polling step for Create or Update, even for "fast" resources.
- Never use `time.Sleep` directly. Always use `GetXxxUntilState()`.
- Never hard-code delay/interval values. Always resolve through `resource.retry.with(data.Retry)`.
- Never ignore the error from `GetXxxUntilState()`.
