package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func TestBlockStorageToDataSourceModel(t *testing.T) {
	createdAt := time.Now()
	modifiedAt := createdAt.Add(1 * time.Hour)
	deletedAt := createdAt.Add(2 * time.Hour)

	block := &sdk.BlockStorage{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Name:           "block-storage-1",
			Workspace:      "workspace-1",
			Tenant:         "tenant-1",
			Region:         "region-1",
			Ref:            "seca.storage/v1/tenants/tenant-1/workspaces/workspace-1/block-storages/block-storage-1",
			CreatedAt:      createdAt,
			DeletedAt:      &deletedAt,
			LastModifiedAt: modifiedAt,
		},
		Labels:      sdk.Labels{"env": "prod"},
		Annotations: sdk.Annotations{"team": "core"},
		Extensions:  sdk.Extensions{"ext": "v1"},
		Spec: sdk.BlockStorageSpec{
			SizeGB: 200,
			SkuRef: sdk.Reference{Resource: "storage-skus/sku-1"},
		},
		Status: &sdk.BlockStorageStatus{State: sdk.ResourceStateActive},
	}

	model, diags := blockStorageToDataSourceModel(context.Background(), block)
	require.False(t, diags.HasError())

	assert.Equal(t, "seca.storage/v1/tenants/tenant-1/workspaces/workspace-1/block-storages/block-storage-1", model.Id.ValueString())
	assert.Equal(t, "block-storage-1", model.Name.ValueString())
	assert.Equal(t, "workspace-1", model.WorkspaceId.ValueString())
	assert.Equal(t, "seca.storage/v1", model.ResourceProvider.ValueString())

	assert.Equal(t, createdAt.Format(time.RFC3339), model.CreatedAt.ValueString())
	assert.Equal(t, deletedAt.Format(time.RFC3339), model.DeletedAt.ValueString())
	assert.Equal(t, modifiedAt.Format(time.RFC3339), model.LastModifiedAt.ValueString())

	assert.Equal(t, map[string]string{"env": "prod"}, toStringMap(model.Labels))
	assert.Equal(t, map[string]string{"team": "core"}, toStringMap(model.Annotations))
	assert.Equal(t, map[string]string{"ext": "v1"}, toStringMap(model.Extensions))

	assert.Equal(t, int64(200), model.SizeGB.ValueInt64())
	assert.Equal(t, "storage-skus/sku-1", model.SkuId.ValueString())
	assert.True(t, model.SourceImageId.IsNull())

	assert.Equal(t, string(sdk.ResourceStateActive), model.State.ValueString())
}
