package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func TestNetworkSkuToDataSourceModel(t *testing.T) {
	sku := &sdk.NetworkSku{
		Metadata: &sdk.SkuResourceMetadata{
			Name:   "N10K",
			Tenant: "tenant-1",
			Region: "region-1",
			Ref:    "seca.network/v1/tenants/tenant-1/network-skus/N10K",
		},
		Labels:      sdk.Labels{"tier": "standard"},
		Annotations: sdk.Annotations{"team": "network"},
		Extensions:  sdk.Extensions{"ext": "v1"},
		Spec: &sdk.NetworkSkuSpec{
			Bandwidth: 10000,
			Packets:   1000000,
		},
	}

	model, diags := networkSkuToDataSourceModel(context.Background(), sku)
	require.False(t, diags.HasError())

	assert.Equal(t, "seca.network/v1/tenants/tenant-1/network-skus/N10K", model.Id.ValueString())
	assert.Equal(t, "N10K", model.Name.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "region-1", model.Region.ValueString())
	assert.Equal(t, "seca.network/v1", model.ResourceProvider.ValueString())

	assert.Equal(t, map[string]string{"tier": "standard"}, toStringMap(model.Labels))
	assert.Equal(t, map[string]string{"team": "network"}, toStringMap(model.Annotations))
	assert.Equal(t, map[string]string{"ext": "v1"}, toStringMap(model.Extensions))

	assert.Equal(t, int64(10000), model.Bandwidth.ValueInt64())
	assert.Equal(t, int64(1000000), model.Packets.ValueInt64())
}

func TestNetworkSkuToDataSourceModel_NilSpec(t *testing.T) {
	sku := &sdk.NetworkSku{
		Metadata: &sdk.SkuResourceMetadata{
			Name:   "N10K",
			Tenant: "tenant-1",
			Region: "region-1",
			Ref:    "seca.network/v1/tenants/tenant-1/network-skus/N10K",
		},
		Spec: nil,
	}

	model, diags := networkSkuToDataSourceModel(context.Background(), sku)
	require.False(t, diags.HasError())

	assert.True(t, model.Bandwidth.IsNull())
	assert.True(t, model.Packets.IsNull())
}
