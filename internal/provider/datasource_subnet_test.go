package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func TestSubnetToDataSourceModel(t *testing.T) {
	createdAt := time.Now()
	modifiedAt := createdAt.Add(1 * time.Hour)
	deletedAt := createdAt.Add(2 * time.Hour)

	sub := &sdk.Subnet{
		Metadata: &sdk.RegionalNetworkResourceMetadata{
			Name:           "subnet-1",
			Workspace:      "workspace-1",
			Network:        "network-1",
			Tenant:         "tenant-1",
			Region:         "region-1",
			Ref:            "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/networks/network-1/subnets/subnet-1",
			CreatedAt:      createdAt,
			DeletedAt:      &deletedAt,
			LastModifiedAt: modifiedAt,
		},
		Labels:      sdk.Labels{"env": "prod"},
		Annotations: sdk.Annotations{"team": "network"},
		Extensions:  sdk.Extensions{"ext": "v1"},
		Spec: sdk.SubnetSpec{
			Cidr:          sdk.Cidr{Ipv4: "10.0.1.0/24"},
			RouteTableRef: sdk.Reference{Resource: "route-tables/rt-1"},
			Zone:          "zone-1",
			SkuRef:        &sdk.Reference{Resource: "network-skus/sku-1"},
		},
		Status: &sdk.SubnetStatus{
			State: sdk.ResourceStateActive,
		},
	}

	model, diags := subnetToDataSourceModel(context.Background(), sub)
	require.False(t, diags.HasError())

	assert.Equal(t, "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/networks/network-1/subnets/subnet-1", model.Id.ValueString())
	assert.Equal(t, "subnet-1", model.Name.ValueString())
	assert.Equal(t, "workspace-1", model.WorkspaceId.ValueString())
	assert.Equal(t, "network-1", model.NetworkId.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "region-1", model.Region.ValueString())
	assert.Equal(t, "seca.network/v1", model.ResourceProvider.ValueString())

	assert.Equal(t, createdAt.Format(time.RFC3339), model.CreatedAt.ValueString())
	assert.Equal(t, deletedAt.Format(time.RFC3339), model.DeletedAt.ValueString())
	assert.Equal(t, modifiedAt.Format(time.RFC3339), model.LastModifiedAt.ValueString())

	assert.Equal(t, map[string]string{"env": "prod"}, toStringMap(model.Labels))
	assert.Equal(t, map[string]string{"team": "network"}, toStringMap(model.Annotations))

	assert.Equal(t, "10.0.1.0/24", model.Cidr.Ipv4.ValueString())
	assert.Equal(t, "route-tables/rt-1", model.RouteTableId.ValueString())
	assert.Equal(t, "zone-1", model.Zone.ValueString())
	assert.Equal(t, "network-skus/sku-1", model.SkuId.ValueString())
	assert.Equal(t, string(sdk.ResourceStateActive), model.State.ValueString())
}

func TestSubnetToDataSourceModel_NilStatus(t *testing.T) {
	sub := &sdk.Subnet{
		Metadata: &sdk.RegionalNetworkResourceMetadata{
			Name:      "subnet-1",
			Workspace: "workspace-1",
			Network:   "network-1",
			Tenant:    "tenant-1",
			Region:    "region-1",
			Ref:       "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/networks/network-1/subnets/subnet-1",
		},
		Spec:   sdk.SubnetSpec{Cidr: sdk.Cidr{Ipv4: "10.0.1.0/24"}},
		Status: nil,
	}

	model, diags := subnetToDataSourceModel(context.Background(), sub)
	require.False(t, diags.HasError())

	assert.True(t, model.State.IsNull())
}
