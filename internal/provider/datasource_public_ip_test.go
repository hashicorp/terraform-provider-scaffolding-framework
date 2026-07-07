package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func TestPublicIpToDataSourceModel(t *testing.T) {
	createdAt := time.Now()
	modifiedAt := createdAt.Add(1 * time.Hour)
	deletedAt := createdAt.Add(2 * time.Hour)

	ip := &sdk.PublicIp{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Name:           "ip-1",
			Workspace:      "workspace-1",
			Tenant:         "tenant-1",
			Region:         "region-1",
			Ref:            "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/public-ips/ip-1",
			CreatedAt:      createdAt,
			DeletedAt:      &deletedAt,
			LastModifiedAt: modifiedAt,
		},
		Labels:      sdk.Labels{"env": "prod"},
		Annotations: sdk.Annotations{"team": "network"},
		Extensions:  sdk.Extensions{"ext": "v1"},
		Spec: sdk.PublicIpSpec{
			Version: sdk.IPVersionIPv4,
		},
		Status: &sdk.PublicIpStatus{
			IpAddress:  "203.0.113.42",
			AttachedTo: &sdk.Reference{Resource: "nics/nic-1"},
			State:      sdk.ResourceStateActive,
		},
	}

	model, diags := publicIpToDataSourceModel(context.Background(), ip)
	require.False(t, diags.HasError())

	assert.Equal(t, "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/public-ips/ip-1", model.Id.ValueString())
	assert.Equal(t, "ip-1", model.Name.ValueString())
	assert.Equal(t, "workspace-1", model.WorkspaceId.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "region-1", model.Region.ValueString())
	assert.Equal(t, "seca.network/v1", model.ResourceProvider.ValueString())

	assert.Equal(t, createdAt.Format(time.RFC3339), model.CreatedAt.ValueString())
	assert.Equal(t, deletedAt.Format(time.RFC3339), model.DeletedAt.ValueString())
	assert.Equal(t, modifiedAt.Format(time.RFC3339), model.LastModifiedAt.ValueString())

	assert.Equal(t, map[string]string{"env": "prod"}, toStringMap(model.Labels))
	assert.Equal(t, map[string]string{"team": "network"}, toStringMap(model.Annotations))

	assert.Equal(t, "IPv4", model.Version.ValueString())
	assert.Equal(t, "203.0.113.42", model.Address.ValueString())
	assert.Equal(t, "nics/nic-1", model.AttachedTo.ValueString())
	assert.Equal(t, string(sdk.ResourceStateActive), model.State.ValueString())
}

func TestPublicIpToDataSourceModel_NilStatus(t *testing.T) {
	ip := &sdk.PublicIp{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Name:      "ip-1",
			Workspace: "workspace-1",
			Tenant:    "tenant-1",
			Region:    "region-1",
			Ref:       "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/public-ips/ip-1",
		},
		Spec:   sdk.PublicIpSpec{Version: sdk.IPVersionIPv4},
		Status: nil,
	}

	model, diags := publicIpToDataSourceModel(context.Background(), ip)
	require.False(t, diags.HasError())

	assert.True(t, model.State.IsNull())
	assert.True(t, model.Address.IsNull())
}
