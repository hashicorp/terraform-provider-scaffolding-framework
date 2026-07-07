package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func securityGroupFixture() *sdk.SecurityGroup {
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	modifiedAt := createdAt.Add(1 * time.Hour)

	return &sdk.SecurityGroup{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Name:           "sg-1",
			Workspace:      "workspace-1",
			Tenant:         "tenant-1",
			Region:         "region-1",
			Ref:            "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/security-groups/sg-1",
			CreatedAt:      createdAt,
			LastModifiedAt: modifiedAt,
		},
		Spec: sdk.SecurityGroupSpec{
			Rules: []sdk.SecurityGroupRuleSpec{
				{
					Direction: sdk.SecurityGroupRuleDirectionIngress,
					Protocol:  sdk.SecurityGroupRuleProtocolTCP,
					Ports: &sdk.Ports{
						List: []int{80, 443},
					},
				},
				{
					Direction: sdk.SecurityGroupRuleDirectionIngress,
					Protocol:  sdk.SecurityGroupRuleProtocolTCP,
					Ports: &sdk.Ports{
						From: 22,
					},
					SourceRef: []sdk.Reference{{Resource: "55.44.33.11"}},
				},
			},
		},
		Status: &sdk.SecurityGroupStatus{
			State: sdk.ResourceStateActive,
		},
	}
}

func TestSecurityGroupToResourceModel(t *testing.T) {
	sg := securityGroupFixture()

	model, diags := securityGroupToResourceModel(context.Background(), sg)
	require.False(t, diags.HasError())

	assert.Equal(t, sg.Metadata.Ref, model.Id.ValueString())
	assert.Equal(t, "sg-1", model.Name.ValueString())
	assert.Equal(t, "workspace-1", model.WorkspaceId.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "seca.network/v1", model.ResourceProvider.ValueString())
	assert.Equal(t, 2, len(model.Rules.Elements()))
	assert.False(t, model.RuleRefs.IsNull())
	assert.Equal(t, 0, len(model.RuleRefs.Elements()))
}

func TestSecurityGroupToResourceModel_EmptyRules(t *testing.T) {
	sg := securityGroupFixture()
	sg.Spec.Rules = nil

	model, diags := securityGroupToResourceModel(context.Background(), sg)
	require.False(t, diags.HasError())

	assert.Equal(t, 0, len(model.Rules.Elements()))
}

func TestSecurityGroupToResourceModel_WithRuleRefs(t *testing.T) {
	sg := securityGroupFixture()
	sg.Spec.RuleRefs = []sdk.Reference{
		{Resource: "security-group-rules/rule-1"},
	}

	model, diags := securityGroupToResourceModel(context.Background(), sg)
	require.False(t, diags.HasError())

	assert.Equal(t, 1, len(model.RuleRefs.Elements()))
}

func TestSecurityGroupFromModel_RoundTrip(t *testing.T) {
	sg := securityGroupFixture()
	sg.Spec.Rules = []sdk.SecurityGroupRuleSpec{
		{
			Direction: sdk.SecurityGroupRuleDirectionIngress,
			Protocol:  sdk.SecurityGroupRuleProtocolTCP,
			Ports: &sdk.Ports{
				List: []int{80, 443},
			},
		},
	}

	ctx := context.Background()
	model, diags := securityGroupToResourceModel(ctx, sg)
	require.False(t, diags.HasError())

	roundTripped, diags := securityGroupFromModel(ctx, "tenant-1", model)
	require.False(t, diags.HasError())

	require.Len(t, roundTripped.Spec.Rules, 1)
	assert.Equal(t, sdk.SecurityGroupRuleDirectionIngress, roundTripped.Spec.Rules[0].Direction)
	assert.Equal(t, sdk.SecurityGroupRuleProtocolTCP, roundTripped.Spec.Rules[0].Protocol)
	require.NotNil(t, roundTripped.Spec.Rules[0].Ports)
	assert.Equal(t, []int{80, 443}, roundTripped.Spec.Rules[0].Ports.List)
}

func TestSecurityGroupPortsFromNilPorts(t *testing.T) {
	obj, diags := sgPortsToObjectValue(context.Background(), nil)
	require.False(t, diags.HasError())
	assert.True(t, obj.IsNull())
}

func TestSecurityGroupPortsFromRange(t *testing.T) {
	ports := &sdk.Ports{From: 22}

	obj, diags := sgPortsToObjectValue(context.Background(), ports)
	require.False(t, diags.HasError())
	assert.False(t, obj.IsNull())
	attrs := obj.Attributes()
	assert.Equal(t, types.Int64Value(22), attrs["from"])
	assert.Equal(t, types.Int64Value(0), attrs["to"])
}
