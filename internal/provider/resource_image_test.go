package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func TestImageToResourceModel(t *testing.T) {
	createdAt := time.Now()
	modifiedAt := createdAt.Add(1 * time.Hour)
	deletedAt := createdAt.Add(2 * time.Hour)

	image := &sdk.Image{
		Metadata: &sdk.RegionalResourceMetadata{
			Name:           "image-1",
			Tenant:         "tenant-1",
			Region:         "region-1",
			Ref:            "images/image-1",
			CreatedAt:      createdAt,
			DeletedAt:      &deletedAt,
			LastModifiedAt: modifiedAt,
		},
		Labels:      sdk.Labels{"env": "prod"},
		Annotations: sdk.Annotations{"team": "core"},
		Extensions:  sdk.Extensions{"ext": "v1"},
		Spec: sdk.ImageSpec{
			BlockStorageRef: sdk.Reference{Resource: "block-storages/block-storage-1"},
			CpuArchitecture: sdk.ImageSpecCpuArchitectureAmd64,
			Initializer:     sdk.ImageSpecInitializerCloudinit22,
			Boot:            sdk.ImageSpecBootUEFI,
		},
	}

	model, diags := imageToResourceModel(context.Background(), image)
	require.False(t, diags.HasError())

	assert.Equal(t, "images/image-1", model.Id.ValueString())
	assert.Equal(t, "image-1", model.Name.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "region-1", model.Region.ValueString())

	assert.Equal(t, createdAt.Format(time.RFC3339), model.CreatedAt.ValueString())
	assert.Equal(t, deletedAt.Format(time.RFC3339), model.DeletedAt.ValueString())
	assert.Equal(t, modifiedAt.Format(time.RFC3339), model.LastModifiedAt.ValueString())

	assert.Equal(t, map[string]string{"env": "prod"}, toStringMap(model.Labels))
	assert.Equal(t, map[string]string{"team": "core"}, toStringMap(model.Annotations))
	assert.Equal(t, map[string]string{"ext": "v1"}, toStringMap(model.Extensions))

	assert.Equal(t, "block-storages/block-storage-1", model.BlockStorageId.ValueString())
	assert.Equal(t, "amd64", model.CpuArchitecture.ValueString())
	assert.Equal(t, "cloudinit-22", model.Initializer.ValueString())
	assert.Equal(t, "UEFI", model.Boot.ValueString())
}
