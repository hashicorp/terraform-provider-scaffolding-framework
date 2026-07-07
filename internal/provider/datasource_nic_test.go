package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func TestNicToDataSourceModel(t *testing.T) {
	createdAt := time.Now()
	modifiedAt := createdAt.Add(1 * time.Hour)
	deletedAt := createdAt.Add(2 * time.Hour)

	nic := &sdk.Nic{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Name:           "nic-1",
			Workspace:      "workspace-1",
			Tenant:         "tenant-1",
			Region:         "region-1",
			Ref:            "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/nics/nic-1",
			CreatedAt:      createdAt,
			DeletedAt:      &deletedAt,
			LastModifiedAt: modifiedAt,
		},
		Labels:      sdk.Labels{"env": "prod"},
		Annotations: sdk.Annotations{"team": "network"},
		Extensions:  sdk.Extensions{"ext": "v1"},
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

	model, diags := nicToDataSourceModel(context.Background(), nic)
	require.False(t, diags.HasError())

	assert.Equal(t, "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/nics/nic-1", model.Id.ValueString())
	assert.Equal(t, "nic-1", model.Name.ValueString())
	assert.Equal(t, "workspace-1", model.WorkspaceId.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "region-1", model.Region.ValueString())
	assert.Equal(t, "seca.network/v1", model.ResourceProvider.ValueString())

	assert.Equal(t, createdAt.Format(time.RFC3339), model.CreatedAt.ValueString())
	assert.Equal(t, deletedAt.Format(time.RFC3339), model.DeletedAt.ValueString())
	assert.Equal(t, modifiedAt.Format(time.RFC3339), model.LastModifiedAt.ValueString())

	assert.Equal(t, map[string]string{"env": "prod"}, toStringMap(model.Labels))
	assert.Equal(t, map[string]string{"team": "network"}, toStringMap(model.Annotations))

	assert.Equal(t, "subnets/subnet-1", model.SubnetId.ValueString())
	assert.Equal(t, 1, len(model.Addresses.Elements()))
	assert.Equal(t, "aa:bb:cc:dd:ee:ff", model.MacAddress.ValueString())
	assert.Equal(t, 1, len(model.PublicIpIds.Elements()))
	assert.Equal(t, "network-skus/sku-1", model.SkuId.ValueString())
	assert.Equal(t, string(sdk.ResourceStateActive), model.State.ValueString())
}

func TestNicToDataSourceModel_NilStatus(t *testing.T) {
	nic := &sdk.Nic{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Name:      "nic-1",
			Workspace: "workspace-1",
			Tenant:    "tenant-1",
			Region:    "region-1",
			Ref:       "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/nics/nic-1",
		},
		Spec: sdk.NicSpec{
			SubnetRef: sdk.Reference{Resource: "subnets/subnet-1"},
		},
		Status: nil,
	}

	model, diags := nicToDataSourceModel(context.Background(), nic)
	require.False(t, diags.HasError())

	assert.True(t, model.State.IsNull())
	assert.True(t, model.MacAddress.IsNull())
}
