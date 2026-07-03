package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func TestRegionToDataSourceModel(t *testing.T) {
	createdAt := time.Now()
	modifiedAt := createdAt.Add(1 * time.Hour)
	deletedAt := createdAt.Add(2 * time.Hour)

	region := &sdk.Region{
		Metadata: &sdk.GlobalResourceMetadata{
			Name:           "region-1",
			Ref:            "seca.region/v1/regions/region-1",
			CreatedAt:      createdAt,
			DeletedAt:      &deletedAt,
			LastModifiedAt: modifiedAt,
		},
		Spec: sdk.RegionSpec{
			AvailableZones: []sdk.Zone{"zone-a", "zone-b"},
			Providers: []sdk.Provider{
				{Name: "seca.workspace", Version: "v1", Url: "http://example.com/workspace"},
				{Name: "seca.storage", Version: "v1", Url: "http://example.com/storage"},
			},
		},
	}

	model, diags := regionToDataSourceModel(context.Background(), region)
	require.False(t, diags.HasError())

	assert.Equal(t, "seca.region/v1/regions/region-1", model.Id.ValueString())
	assert.Equal(t, "region-1", model.Name.ValueString())
	assert.Equal(t, "seca.region/v1", model.ResourceProvider.ValueString())

	assert.Equal(t, createdAt.Format(time.RFC3339), model.CreatedAt.ValueString())
	assert.Equal(t, deletedAt.Format(time.RFC3339), model.DeletedAt.ValueString())
	assert.Equal(t, modifiedAt.Format(time.RFC3339), model.LastModifiedAt.ValueString())

	var zones []string
	diags = model.AvailableZones.ElementsAs(context.Background(), &zones, false)
	require.False(t, diags.HasError())
	assert.Equal(t, []string{"zone-a", "zone-b"}, zones)

	var providers []RegionProviderModel
	diags = model.Providers.ElementsAs(context.Background(), &providers, false)
	require.False(t, diags.HasError())
	require.Len(t, providers, 2)
	assert.Equal(t, "seca.workspace", providers[0].Name.ValueString())
	assert.Equal(t, "v1", providers[0].Version.ValueString())
	assert.Equal(t, "http://example.com/workspace", providers[0].Url.ValueString())
}
