package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func nicFixture() *sdk.Nic {
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	modifiedAt := createdAt.Add(1 * time.Hour)

	return &sdk.Nic{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Name:           "nic-1",
			Workspace:      "workspace-1",
			Tenant:         "tenant-1",
			Region:         "region-1",
			Ref:            "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/nics/nic-1",
			CreatedAt:      createdAt,
			LastModifiedAt: modifiedAt,
		},
		Spec: sdk.NicSpec{
			SubnetRef: sdk.Reference{Resource: "subnets/subnet-1"},
			Addresses: []string{"10.0.1.10"},
			SkuRef:    &sdk.Reference{Resource: "network-skus/sku-1"},
		},
		Status: &sdk.NicStatus{
			MacAddress:   "aa:bb:cc:dd:ee:ff",
			PublicIpRefs: []sdk.Reference{{Resource: "public-ips/ip-1"}},
			State:        sdk.ResourceStateActive,
		},
	}
}

func TestNicToResourceModel(t *testing.T) {
	nic := nicFixture()

	model, diags := nicToResourceModel(context.Background(), nic)
	require.False(t, diags.HasError())

	assert.Equal(t, nic.Metadata.Ref, model.Id.ValueString())
	assert.Equal(t, "nic-1", model.Name.ValueString())
	assert.Equal(t, "workspace-1", model.WorkspaceId.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "seca.network/v1", model.ResourceProvider.ValueString())
	assert.Equal(t, "subnets/subnet-1", model.SubnetId.ValueString())
	assert.Equal(t, 1, len(model.Addresses.Elements()))
	assert.Equal(t, "aa:bb:cc:dd:ee:ff", model.MacAddress.ValueString())
	assert.Equal(t, "network-skus/sku-1", model.SkuId.ValueString())
	assert.Equal(t, 1, len(model.PublicIpIds.Elements()))
}

func TestNicToResourceModel_NilStatus(t *testing.T) {
	nic := nicFixture()
	nic.Status = nil

	model, diags := nicToResourceModel(context.Background(), nic)
	require.False(t, diags.HasError())

	assert.True(t, model.MacAddress.IsNull())
	assert.Equal(t, 0, len(model.PublicIpIds.Elements()))
}

func TestNicToResourceModel_EmptyAddresses(t *testing.T) {
	nic := nicFixture()
	nic.Spec.Addresses = nil

	model, diags := nicToResourceModel(context.Background(), nic)
	require.False(t, diags.HasError())

	assert.Equal(t, 0, len(model.Addresses.Elements()))
}

func TestNicFromModel_RoundTrip(t *testing.T) {
	nic := nicFixture()

	ctx := context.Background()
	model, diags := nicToResourceModel(ctx, nic)
	require.False(t, diags.HasError())

	roundTripped, diags := nicFromModel(ctx, "tenant-1", model)
	require.False(t, diags.HasError())

	assert.Equal(t, "subnets/subnet-1", roundTripped.Spec.SubnetRef.Resource)
	assert.Equal(t, []string{"10.0.1.10"}, roundTripped.Spec.Addresses)
}

func TestNicToDataSourceModel(t *testing.T) {
	nic := nicFixture()

	model, diags := nicToDataSourceModel(context.Background(), nic)
	require.False(t, diags.HasError())

	assert.Equal(t, nic.Metadata.Ref, model.Id.ValueString())
	assert.Equal(t, "subnets/subnet-1", model.SubnetId.ValueString())
	assert.Equal(t, "aa:bb:cc:dd:ee:ff", model.MacAddress.ValueString())
	assert.Equal(t, string(sdk.ResourceStateActive), model.State.ValueString())
}

func TestNicToDataSourceModel_NilStatus(t *testing.T) {
	nic := nicFixture()
	nic.Status = nil

	model, diags := nicToDataSourceModel(context.Background(), nic)
	require.False(t, diags.HasError())

	assert.True(t, model.State.IsNull())
	assert.True(t, model.MacAddress.IsNull())
}
