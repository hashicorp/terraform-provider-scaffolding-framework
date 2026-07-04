package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func subnetFixture() *sdk.Subnet {
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	modifiedAt := createdAt.Add(1 * time.Hour)

	return &sdk.Subnet{
		Metadata: &sdk.RegionalNetworkResourceMetadata{
			Name:           "subnet-1",
			Workspace:      "workspace-1",
			Network:        "network-1",
			Tenant:         "tenant-1",
			Region:         "region-1",
			Ref:            "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/networks/network-1/subnets/subnet-1",
			CreatedAt:      createdAt,
			LastModifiedAt: modifiedAt,
		},
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
}

func TestSubnetToResourceModel(t *testing.T) {
	sub := subnetFixture()

	model, diags := subnetToResourceModel(context.Background(), sub)
	require.False(t, diags.HasError())

	assert.Equal(t, sub.Metadata.Ref, model.Id.ValueString())
	assert.Equal(t, "subnet-1", model.Name.ValueString())
	assert.Equal(t, "workspace-1", model.WorkspaceId.ValueString())
	assert.Equal(t, "network-1", model.NetworkId.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "region-1", model.Region.ValueString())
	assert.Equal(t, "seca.network/v1", model.ResourceProvider.ValueString())
	assert.Equal(t, "10.0.1.0/24", model.Cidr.Ipv4.ValueString())
	assert.Equal(t, "route-tables/rt-1", model.RouteTableId.ValueString())
	assert.Equal(t, "zone-1", model.Zone.ValueString())
	assert.Equal(t, "network-skus/sku-1", model.SkuId.ValueString())
}

func TestSubnetToResourceModel_NilRouteTable(t *testing.T) {
	sub := subnetFixture()
	sub.Spec.RouteTableRef = sdk.Reference{}

	model, diags := subnetToResourceModel(context.Background(), sub)
	require.False(t, diags.HasError())

	assert.Equal(t, "", model.RouteTableId.ValueString())
}

func TestSubnetFromModel_RoundTrip(t *testing.T) {
	sub := subnetFixture()

	ctx := context.Background()
	model, diags := subnetToResourceModel(ctx, sub)
	require.False(t, diags.HasError())

	roundTripped := subnetFromModel("tenant-1", model)
	assert.Equal(t, "10.0.1.0/24", roundTripped.Spec.Cidr.Ipv4)
	assert.Equal(t, "route-tables/rt-1", roundTripped.Spec.RouteTableRef.Resource)
}

func TestSubnetFromModel_NilRouteTableId(t *testing.T) {
	sub := subnetFixture()
	sub.Spec.RouteTableRef = sdk.Reference{}

	ctx := context.Background()
	model, diags := subnetToResourceModel(ctx, sub)
	require.False(t, diags.HasError())

	roundTripped := subnetFromModel("tenant-1", model)
	assert.Empty(t, roundTripped.Spec.RouteTableRef.Resource)
}

func TestSubnetToDataSourceModel(t *testing.T) {
	sub := subnetFixture()

	model, diags := subnetToDataSourceModel(context.Background(), sub)
	require.False(t, diags.HasError())

	assert.Equal(t, sub.Metadata.Ref, model.Id.ValueString())
	assert.Equal(t, "network-1", model.NetworkId.ValueString())
	assert.Equal(t, string(sdk.ResourceStateActive), model.State.ValueString())
	assert.False(t, model.Cidr.IsNull())
}

func TestSubnetToDataSourceModel_NilStatus(t *testing.T) {
	sub := subnetFixture()
	sub.Status = nil

	model, diags := subnetToDataSourceModel(context.Background(), sub)
	require.False(t, diags.HasError())

	assert.True(t, model.State.IsNull())
}
