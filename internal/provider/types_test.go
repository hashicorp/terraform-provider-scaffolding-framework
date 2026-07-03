package provider

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func TestNumberToDuration(t *testing.T) {
	tests := []struct {
		name string
		in   types.Number
		want time.Duration
	}{
		{"null", types.NumberNull(), 0},
		{"unknown", types.NumberUnknown(), 0},
		{"zero", types.NumberValue(big.NewFloat(0)), 0},
		{"whole seconds", types.NumberValue(big.NewFloat(30)), 30 * time.Second},
		{"fractional seconds", types.NumberValue(big.NewFloat(1.5)), 1500 * time.Millisecond},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, numberToDuration(tt.in))
		})
	}
}

func TestNumberToInt(t *testing.T) {
	tests := []struct {
		name string
		in   types.Number
		want int
	}{
		{"null", types.NumberNull(), 0},
		{"unknown", types.NumberUnknown(), 0},
		{"zero", types.NumberValue(big.NewFloat(0)), 0},
		{"positive", types.NumberValue(big.NewFloat(42)), 42},
		{"negative", types.NumberValue(big.NewFloat(-7)), -7},
		{"truncates fraction", types.NumberValue(big.NewFloat(9.9)), 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, numberToInt(tt.in))
		})
	}
}

func TestToStringMap(t *testing.T) {
	t.Run("null", func(t *testing.T) {
		assert.Nil(t, toStringMap(types.MapNull(types.StringType)))
	})

	t.Run("unknown", func(t *testing.T) {
		assert.Nil(t, toStringMap(types.MapUnknown(types.StringType)))
	})

	t.Run("empty", func(t *testing.T) {
		m, diags := types.MapValueFrom(context.Background(), types.StringType, map[string]string{})
		require.False(t, diags.HasError())
		assert.Empty(t, toStringMap(m))
	})

	t.Run("populated", func(t *testing.T) {
		want := map[string]string{"env": "prod", "team": "core"}
		m, diags := types.MapValueFrom(context.Background(), types.StringType, want)
		require.False(t, diags.HasError())
		assert.Equal(t, want, toStringMap(m))
	})
}

func TestFromStringMap(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		m, diags := fromStringMap(context.Background(), nil)
		require.False(t, diags.HasError())
		assert.True(t, m.IsNull())
	})

	t.Run("empty", func(t *testing.T) {
		m, diags := fromStringMap(context.Background(), map[string]string{})
		require.False(t, diags.HasError())
		assert.True(t, m.IsNull())
	})

	t.Run("populated", func(t *testing.T) {
		in := map[string]string{"env": "prod", "team": "core"}
		m, diags := fromStringMap(context.Background(), in)
		require.False(t, diags.HasError())
		assert.False(t, m.IsNull())
		assert.Equal(t, in, toStringMap(m))
	})
}

func TestFromTime(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		assert.True(t, fromTime(time.Time{}).IsNull())
	})

	t.Run("value", func(t *testing.T) {
		ts := time.Date(2026, 6, 23, 10, 30, 0, 0, time.UTC)
		got := fromTime(ts)
		assert.False(t, got.IsNull())
		assert.Equal(t, ts.Format(time.RFC3339), got.ValueString())
	})
}

func TestFromTimePtr(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.True(t, fromTimePtr(nil).IsNull())
	})

	t.Run("zero value", func(t *testing.T) {
		ts := time.Time{}
		assert.True(t, fromTimePtr(&ts).IsNull())
	})

	t.Run("value", func(t *testing.T) {
		ts := time.Date(2026, 6, 23, 10, 30, 0, 0, time.UTC)
		got := fromTimePtr(&ts)
		assert.False(t, got.IsNull())
		assert.Equal(t, ts.Format(time.RFC3339), got.ValueString())
	})
}

func TestFromRefPtr(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.True(t, fromRefPtr(nil).IsNull())
	})

	t.Run("value", func(t *testing.T) {
		ref := &sdk.Reference{Resource: "some-resource"}
		got := fromRefPtr(ref)
		assert.False(t, got.IsNull())
		assert.Equal(t, "some-resource", got.ValueString())
	})
}

func TestRefToResourceProvider(t *testing.T) {
	tests := []struct {
		name string
		ref  string
		want string
		null bool
	}{
		{
			name: "empty string",
			ref:  "",
			null: true,
		},
		{
			name: "workspace ref",
			ref:  "seca.workspace/v1/tenants/tenant-1/workspaces/workspace-1",
			want: "seca.workspace/v1",
		},
		{
			name: "storage ref (image)",
			ref:  "seca.storage/v1/tenants/tenant-1/images/image-1",
			want: "seca.storage/v1",
		},
		{
			name: "storage ref (block storage)",
			ref:  "seca.storage/v1/tenants/tenant-1/workspaces/workspace-1/block-storages/bs-1",
			want: "seca.storage/v1",
		},
		{
			name: "region ref",
			ref:  "seca.region/v1/regions/region-1",
			want: "seca.region/v1",
		},
		{
			name: "storage sku ref",
			ref:  "seca.storage/v1/tenants/tenant-1/storage-skus/RD500",
			want: "seca.storage/v1",
		},
		{
			name: "no slash — returns as-is",
			ref:  "noslash",
			want: "noslash",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := refToResourceProvider(tt.ref)
			if tt.null {
				assert.True(t, got.IsNull())
			} else {
				assert.False(t, got.IsNull())
				assert.Equal(t, tt.want, got.ValueString())
			}
		})
	}
}
