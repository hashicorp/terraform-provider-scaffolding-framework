package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func publicIpFixture() *sdk.PublicIp {
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	modifiedAt := createdAt.Add(1 * time.Hour)

	return &sdk.PublicIp{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Name:           "ip-1",
			Workspace:      "workspace-1",
			Tenant:         "tenant-1",
			Region:         "region-1",
			Ref:            "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/public-ips/ip-1",
			CreatedAt:      createdAt,
			LastModifiedAt: modifiedAt,
		},
		Spec: sdk.PublicIpSpec{
			Version: sdk.IPVersionIPv4,
		},
		Status: &sdk.PublicIpStatus{
			IpAddress:  "203.0.113.42",
			AttachedTo: &sdk.Reference{Resource: "nics/nic-1"},
			State:      sdk.ResourceStateActive,
		},
	}
}

func TestPublicIpToResourceModel(t *testing.T) {
	ip := publicIpFixture()

	model, diags := publicIpToResourceModel(context.Background(), ip)
	require.False(t, diags.HasError())

	assert.Equal(t, ip.Metadata.Ref, model.Id.ValueString())
	assert.Equal(t, "ip-1", model.Name.ValueString())
	assert.Equal(t, "workspace-1", model.WorkspaceId.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "seca.network/v1", model.ResourceProvider.ValueString())
	assert.Equal(t, "IPv4", model.Version.ValueString())
	assert.Equal(t, "203.0.113.42", model.Address.ValueString())
	assert.Equal(t, "203.0.113.42", model.IpAddress.ValueString())
	assert.Equal(t, "nics/nic-1", model.AttachedTo.ValueString())
}

func TestPublicIpToResourceModel_NilStatus(t *testing.T) {
	ip := publicIpFixture()
	ip.Status = nil

	model, diags := publicIpToResourceModel(context.Background(), ip)
	require.False(t, diags.HasError())

	assert.True(t, model.Address.IsNull())
	assert.True(t, model.IpAddress.IsNull())
	assert.True(t, model.AttachedTo.IsNull())
}

func TestPublicIpFromModel_RoundTrip(t *testing.T) {
	ip := publicIpFixture()

	ctx := context.Background()
	model, diags := publicIpToResourceModel(ctx, ip)
	require.False(t, diags.HasError())

	roundTripped := publicIpFromModel("tenant-1", model)
	assert.Equal(t, sdk.IPVersionIPv4, roundTripped.Spec.Version)
	assert.Equal(t, "workspace-1", roundTripped.Metadata.Workspace)
}

