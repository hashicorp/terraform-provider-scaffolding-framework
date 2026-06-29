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
config := secapi.ResourceObserverUntilValueConfig[sdk.ResourceState]{
    ExpectedValues: []sdk.ResourceState{sdk.ResourceStateActive},
    Delay:          resource.retryDelay,       // time before first poll
    Interval:       resource.retryInterval,    // time between polls
    MaxAttempts:    resource.retryMaxAttempts, // max number of polls
}

// Step 4: Poll until active (or fail)
result, err = resource.client.StorageV1.GetXxxUntilState(ctx, ref, config)
if err != nil {
    resp.Diagnostics.AddError("Error creating Xxx", "...\nError: "+err.Error())
    return
}
```

**The result from Step 4 (not Step 1) must be used to populate Terraform state.** The state must reflect the fully provisioned resource, not the intermediate response from the create call.

## Retry Configuration

Retry parameters are configured at the provider level and passed to each resource via the `clients` struct:

| Parameter | Default | Provider attribute |
|---|---|---|
| `RetryDelay` | `30s` | `retry.delay` (seconds) |
| `RetryInterval` | `10s` | `retry.interval` (seconds) |
| `RetryMaxAttempts` | `5` | `retry.max_attempts` |

`Delay` is the initial wait before the first poll (allows the API time to begin provisioning). `Interval` is the wait between subsequent polls.

These can be overridden per provider block:

```hcl
provider "seca" {
  retry = {
    delay        = 60
    interval     = 15
    max_attempts = 20
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

`Delete()` does **not** poll. It calls `DeleteXxx()` and returns immediately. This is a known gap (see [known-issues.md](known-issues.md)). The API is expected to handle deletion asynchronously, and subsequent `Read()` calls will detect when the resource is gone.

## What NOT to Do

- Never write state from the Step 1 result. Always use the Step 4 (poll) result.
- Never skip the polling step for Create or Update, even for "fast" resources.
- Never use `time.Sleep` directly. Always use `GetXxxUntilState()`.
- Never hard-code delay/interval values. Always use the resource struct's retry fields.
- Never ignore the error from `GetXxxUntilState()`.
