package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func TestWorkspaceToResourceModel(t *testing.T) {
	createdAt := time.Now()
	modifiedAt := createdAt.Add(1 * time.Hour)
	deletedAt := createdAt.Add(2 * time.Hour)

	workspace := &sdk.Workspace{
		Metadata: &sdk.RegionalResourceMetadata{
			Name:           "workspace-1",
			Tenant:         "tenant-1",
			Region:         "region-1",
			Ref:            "seca.workspace/v1/tenants/tenant-1/workspaces/workspace-1",
			CreatedAt:      createdAt,
			DeletedAt:      &deletedAt,
			LastModifiedAt: modifiedAt,
		},
		Labels:      sdk.Labels{"env": "prod"},
		Annotations: sdk.Annotations{"team": "core"},
		Extensions:  sdk.Extensions{"ext": "v1"},
	}

	model, diags := workspaceToResourceModel(context.Background(), workspace)
	require.False(t, diags.HasError())

	assert.Equal(t, "seca.workspace/v1/tenants/tenant-1/workspaces/workspace-1", model.Id.ValueString())
	assert.Equal(t, "workspace-1", model.Name.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "region-1", model.Region.ValueString())
	assert.Equal(t, "seca.workspace/v1", model.ResourceProvider.ValueString())

	assert.Equal(t, createdAt.Format(time.RFC3339), model.CreatedAt.ValueString())
	assert.Equal(t, deletedAt.Format(time.RFC3339), model.DeletedAt.ValueString())
	assert.Equal(t, modifiedAt.Format(time.RFC3339), model.LastModifiedAt.ValueString())

	assert.Equal(t, map[string]string{"env": "prod"}, toStringMap(model.Labels))
	assert.Equal(t, map[string]string{"team": "core"}, toStringMap(model.Annotations))
	assert.Equal(t, map[string]string{"ext": "v1"}, toStringMap(model.Extensions))
}
