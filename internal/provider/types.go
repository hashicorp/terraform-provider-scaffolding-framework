package provider

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func numberToDuration(n types.Number) time.Duration {
	if n.IsNull() || n.IsUnknown() {
		return 0
	}
	seconds, _ := n.ValueBigFloat().Float64()
	return time.Duration(seconds * float64(time.Second))
}

func numberToInt(n types.Number) int {
	if n.IsNull() || n.IsUnknown() {
		return 0
	}
	value, _ := n.ValueBigFloat().Int64()
	return int(value)
}

func toStringMap(m types.Map) map[string]string {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	result := make(map[string]string, len(m.Elements()))
	for k, v := range m.Elements() {
		if s, ok := v.(types.String); ok {
			result[k] = s.ValueString()
		}
	}
	return result
}

func fromStringMap(ctx context.Context, m map[string]string) (types.Map, diag.Diagnostics) {
	if len(m) == 0 {
		return types.MapNull(types.StringType), nil
	}
	return types.MapValueFrom(ctx, types.StringType, m)
}

func fromTime(t time.Time) types.String {
	if t.IsZero() {
		return types.StringNull()
	}
	return types.StringValue(t.Format(time.RFC3339))
}

func fromTimePtr(t *time.Time) types.String {
	if t == nil {
		return types.StringNull()
	}
	return fromTime(*t)
}

func fromRefPtr(ref *sdk.Reference) types.String {
	if ref == nil {
		return types.StringNull()
	}
	return types.StringValue(ref.Resource)
}

// refToResourceProvider extracts the "{provider}/{version}" prefix from a
// full resource Ref URN (e.g. "seca.storage/v1" from
// "seca.storage/v1/tenants/t1/workspaces/w1/block-storages/bs1").
func refToResourceProvider(ref string) types.String {
	if ref == "" {
		return types.StringNull()
	}
	parts := strings.SplitN(ref, "/", 3)
	if len(parts) < 2 {
		return types.StringValue(ref)
	}
	return types.StringValue(parts[0] + "/" + parts[1])
}
