# Testing Strategy

## Two-Layer Approach

The provider uses two completely separate test layers, each with a different purpose and location.

## Layer 1: Unit Tests (`internal/provider/`)

**Package:** `package provider` (white-box access to unexported functions)

**Purpose:** Verify that mapping functions between SDK types and Terraform model types work correctly — including all edge cases (null, zero, empty, missing optional fields).

**Scope:** Only `types.go` helpers and the `xxxToXxxModel()` / `xxxFromModel()` private functions. No HTTP calls. No mocking framework.

**Pattern:**
```go
func TestBlockStorageToResourceModel(t *testing.T) {
    // Build an SDK object with known fields
    block := &sdk.BlockStorage{
        Metadata: &sdk.RegionalWorkspaceResourceMetadata{...},
        Spec: sdk.BlockStorageSpec{...},
    }

    // Call the mapping function
    model, diags := blockStorageToResourceModel(context.Background(), block)

    // Assert no errors
    require.False(t, diags.HasError())

    // Assert every field
    assert.Equal(t, "expected", model.Field.ValueString())
    assert.True(t, model.NullableField.IsNull())
}
```

**Test naming:** `Test<FunctionName>` e.g. `TestBlockStorageToResourceModel`, `TestFromTime`.

**What to cover in unit tests:**
- All standard metadata fields (id, name, tenant, region, created_at, deleted_at, last_modified_at)
- All labels/annotations/extensions (non-nil and nil/empty)
- All spec fields (required and optional)
- Optional pointer fields: nil → null, non-nil → value
- Status fields (data sources only): each possible state value
- Edge cases: zero time, nil pointer, empty map, null map

**What NOT to cover in unit tests:** API calls, provider lifecycle, Terraform plan/apply — that's Layer 2.

## Layer 2: Acceptance Tests (`internal/acctest/`)

**Package:** `package acctest` (separate package, no white-box access)

**Purpose:** Verify end-to-end behavior against a live SECA cluster. Tests create real resources, verify their state via Terraform, and destroy them on teardown.

**Guard:** Tests only run when `TF_ACC=1` is set. The CI job sets this. Local runs require a live cluster.

**Cluster:** Tests use the provider config in `provider_test.go`, which hardcodes endpoints at `172.18.0.2:30081`. Acceptance tests cannot currently be run against an arbitrary cluster without editing this file.

**Pattern:**
```go
func TestAccBlockStorage(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccBlockStorageResourceConfig(),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("seca_block_storage.test", "name", "block-storage-1"),
                    ...
                ),
            },
            {
                Config: testAccBlockStorageDataSourceConfig(),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("data.seca_block_storage.test", "state", "active"),
                    ...
                ),
            },
        },
    })
}
```

**Config builder convention:** Private functions named `testAcc<Resource><Type>Config()` where type is `ResourceConfig`, `DataSourceConfig`, or `UpdateConfig`.

**What each acceptance test must cover:**

| Step | What to verify |
|---|---|
| Step 1: Create resource | All user-specified fields are set correctly |
| Step 2: Create + data source | Data source reads resource state; `state` = "active" |
| Step 3 (if updatable fields exist): Update | Updated fields are reflected |

**Missing acceptance test coverage (gaps):**
- No import state tests (`ImportStateVerify: true`)
- No destroy verification (`CheckDestroy`)
- No tests for invalid configurations (expect planning errors)

## Running Tests

```bash
# Unit tests only (no TF_ACC needed)
go test -v -cover -timeout=120s -parallel=10 ./...

# Single unit test
go test -v -run TestBlockStorageToResourceModel ./internal/provider/

# Acceptance tests (requires TF_ACC=1 and live cluster)
TF_ACC=1 go test -v -cover -timeout 120m ./...

# Single acceptance test
TF_ACC=1 go test -v -run TestAccBlockStorage ./internal/acctest/
```

## Mocking Strategy

**There is no mocking.** Unit tests test pure mapping functions that have no external dependencies. Acceptance tests hit a real cluster. This is intentional:

- Mocking the SDK would couple tests to internal SDK implementation details
- The only logic worth unit-testing is the `model ↔ SDK` mapping — and that has no side effects

Do not introduce mocking frameworks (mockery, gomock, testify/mock) for this layer.

## Test Dependencies

- `github.com/stretchr/testify` — `assert` and `require`
- `github.com/hashicorp/terraform-plugin-testing` — acceptance test helpers

Do not use `terraform-plugin-sdk/v2/helper/acctest` or `terraform-plugin-sdk/v2/helper/resource` — depguard blocks them.

## What to Test When Implementing a New Resource

**Unit tests to write:**
1. `TestXxxToResourceModel` — covers the SDK→resource model mapping with a fully populated SDK object and all nullable/optional fields
2. `TestXxxToDataSourceModel` — same but for the data source model, with Status fields
3. (If `xxxFromModel` has conditional logic) — test each conditional branch

**Acceptance tests to write:**
1. `TestAccXxx` with at least:
   - Step 1: Create with all required fields; check all output attributes
   - Step 2: Add data source; verify it reads the created resource correctly
