package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func TestStorageSkuToDataSourceModel(t *testing.T) {
	sku := &sdk.StorageSku{
		Metadata: &sdk.SkuResourceMetadata{
			Name:   "sku-1",
			Tenant: "tenant-1",
			Region: "region-1",
			Ref:    "seca.storage/v1/tenants/tenant-1/storage-skus/sku-1",
		},
		Labels:      sdk.Labels{"tier": "gold"},
		Annotations: sdk.Annotations{"team": "core"},
		Extensions:  sdk.Extensions{"ext": "v1"},
		Spec: &sdk.StorageSkuSpec{
			Iops:          5000,
			Type:          sdk.StorageSkuTypeRemoteDurable,
			MinVolumeSize: 10,
		},
	}

	model, diags := storageSkuToDataSourceModel(context.Background(), sku)
	require.False(t, diags.HasError())

	assert.Equal(t, "seca.storage/v1/tenants/tenant-1/storage-skus/sku-1", model.Id.ValueString())
	assert.Equal(t, "sku-1", model.Name.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "region-1", model.Region.ValueString())
	assert.Equal(t, "seca.storage/v1", model.ResourceProvider.ValueString())

	assert.Equal(t, map[string]string{"tier": "gold"}, toStringMap(model.Labels))
	assert.Equal(t, map[string]string{"team": "core"}, toStringMap(model.Annotations))
	assert.Equal(t, map[string]string{"ext": "v1"}, toStringMap(model.Extensions))

	assert.Equal(t, int64(5000), model.Iops.ValueInt64())
	assert.Equal(t, "remote-durable", model.Type.ValueString())
	assert.Equal(t, int64(10), model.MinVolumeSize.ValueInt64())
}
