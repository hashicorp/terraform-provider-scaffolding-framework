package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func routeTableFixture(routes []sdk.RouteSpec) *sdk.RouteTable {
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	modifiedAt := createdAt.Add(1 * time.Hour)

	return &sdk.RouteTable{
		Metadata: &sdk.RegionalNetworkResourceMetadata{
			Name:           "rt-1",
			Workspace:      "workspace-1",
			Network:        "network-1",
			Tenant:         "tenant-1",
			Region:         "region-1",
			Ref:            "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/networks/network-1/route-tables/rt-1",
			CreatedAt:      createdAt,
			LastModifiedAt: modifiedAt,
		},
		Spec: sdk.RouteTableSpec{
			Routes: routes,
		},
		Status: &sdk.RouteTableStatus{
			State: sdk.ResourceStateActive,
		},
	}
}

func TestRouteTableToResourceModel(t *testing.T) {
	rt := routeTableFixture([]sdk.RouteSpec{
		{
			DestinationCidrBlock: "0.0.0.0/0",
			TargetRef:            sdk.Reference{Resource: "internet-gateways/igw-1"},
		},
	})

	model, diags := routeTableToResourceModel(context.Background(), rt)
	require.False(t, diags.HasError())

	assert.Equal(t, rt.Metadata.Ref, model.Id.ValueString())
	assert.Equal(t, "rt-1", model.Name.ValueString())
	assert.Equal(t, "workspace-1", model.WorkspaceId.ValueString())
	assert.Equal(t, "network-1", model.NetworkId.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "region-1", model.Region.ValueString())
	assert.Equal(t, "seca.network/v1", model.ResourceProvider.ValueString())
	assert.False(t, model.Routes.IsNull())
	assert.Equal(t, 1, len(model.Routes.Elements()))
}

func TestRouteTableToResourceModel_EmptyRoutes(t *testing.T) {
	rt := routeTableFixture(nil)

	model, diags := routeTableToResourceModel(context.Background(), rt)
	require.False(t, diags.HasError())

	assert.False(t, model.Routes.IsNull())
	assert.Equal(t, 0, len(model.Routes.Elements()))
}

func TestRouteTableFromModel_RoundTrip(t *testing.T) {
	rt := routeTableFixture([]sdk.RouteSpec{
		{
			DestinationCidrBlock: "0.0.0.0/0",
			TargetRef:            sdk.Reference{Resource: "internet-gateways/igw-1"},
		},
	})

	ctx := context.Background()
	model, diags := routeTableToResourceModel(ctx, rt)
	require.False(t, diags.HasError())

	roundTripped, diags := routeTableFromModel(ctx, "tenant-1", model)
	require.False(t, diags.HasError())

	require.Len(t, roundTripped.Spec.Routes, 1)
	assert.Equal(t, "0.0.0.0/0", roundTripped.Spec.Routes[0].DestinationCidrBlock)
	assert.Equal(t, "internet-gateways/igw-1", roundTripped.Spec.Routes[0].TargetRef.Resource)
}

func TestRouteTableToDataSourceModel(t *testing.T) {
	rt := routeTableFixture([]sdk.RouteSpec{
		{
			DestinationCidrBlock: "10.0.0.0/8",
			TargetRef:            sdk.Reference{Resource: "instances/inst-1"},
		},
	})

	model, diags := routeTableToDataSourceModel(context.Background(), rt)
	require.False(t, diags.HasError())

	assert.Equal(t, rt.Metadata.Ref, model.Id.ValueString())
	assert.Equal(t, "network-1", model.NetworkId.ValueString())
	assert.Equal(t, string(sdk.ResourceStateActive), model.State.ValueString())
	assert.Equal(t, 1, len(model.Routes.Elements()))
}

func TestRouteTableToDataSourceModel_NilStatus(t *testing.T) {
	rt := routeTableFixture(nil)
	rt.Status = nil

	model, diags := routeTableToDataSourceModel(context.Background(), rt)
	require.False(t, diags.HasError())

	assert.True(t, model.State.IsNull())
}
