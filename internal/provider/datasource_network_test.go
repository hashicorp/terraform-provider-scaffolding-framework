package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func TestNetworkToDataSourceModel(t *testing.T) {
	createdAt := time.Now()
	modifiedAt := createdAt.Add(1 * time.Hour)
	deletedAt := createdAt.Add(2 * time.Hour)

	net := &sdk.Network{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Name:           "network-1",
			Workspace:      "workspace-1",
			Tenant:         "tenant-1",
			Region:         "region-1",
			Ref:            "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/networks/network-1",
			CreatedAt:      createdAt,
			DeletedAt:      &deletedAt,
			LastModifiedAt: modifiedAt,
		},
		Labels:      sdk.Labels{"env": "prod"},
		Annotations: sdk.Annotations{"team": "network"},
		Extensions:  sdk.Extensions{"ext": "v1"},
		Spec: sdk.NetworkSpec{
			SkuRef: sdk.Reference{Resource: "network-skus/N10K"},
			Cidr:   sdk.Cidr{Ipv4: "10.100.0.0/16"},
		},
		Status: &sdk.NetworkStatus{
			State: sdk.ResourceStateActive,
			Cidr:  sdk.Cidr{Ipv4: "10.100.0.0/16"},
		},
	}

	model, diags := networkToDataSourceModel(context.Background(), net)
	require.False(t, diags.HasError())

	assert.Equal(t, "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/networks/network-1", model.Id.ValueString())
	assert.Equal(t, "network-1", model.Name.ValueString())
	assert.Equal(t, "workspace-1", model.WorkspaceId.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "region-1", model.Region.ValueString())
	assert.Equal(t, "seca.network/v1", model.ResourceProvider.ValueString())

	assert.Equal(t, createdAt.Format(time.RFC3339), model.CreatedAt.ValueString())
	assert.Equal(t, deletedAt.Format(time.RFC3339), model.DeletedAt.ValueString())
	assert.Equal(t, modifiedAt.Format(time.RFC3339), model.LastModifiedAt.ValueString())

	assert.Equal(t, map[string]string{"env": "prod"}, toStringMap(model.Labels))
	assert.Equal(t, map[string]string{"team": "network"}, toStringMap(model.Annotations))

	assert.Equal(t, "network-skus/N10K", model.SkuId.ValueString())
	assert.Equal(t, "10.100.0.0/16", model.Cidr.Ipv4.ValueString())
	assert.Equal(t, string(sdk.ResourceStateActive), model.State.ValueString())
}
