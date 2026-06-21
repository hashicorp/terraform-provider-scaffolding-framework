package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
