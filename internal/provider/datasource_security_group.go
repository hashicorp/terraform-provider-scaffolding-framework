package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	tfschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
	"github.com/eu-sovereign-cloud/go-sdk/secapi"
)

var (
	_ datasource.DataSource              = (*SecurityGroupDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*SecurityGroupDataSource)(nil)
)

type SecurityGroupDataSource struct {
	client *secapi.RegionalClient
	tenant string
}

func newSecurityGroupDataSource() datasource.DataSource {
	return &SecurityGroupDataSource{}
}

func (d *SecurityGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

type SecurityGroupDataSourceModel struct {
	securityGroupModel

	State types.String `tfsdk:"state"`
}

func (d *SecurityGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	portsAttrs := map[string]tfschema.Attribute{
		"from": tfschema.Int64Attribute{Computed: true},
		"to":   tfschema.Int64Attribute{Computed: true},
		"list": tfschema.ListAttribute{
			ElementType: types.Int64Type,
			Computed:    true,
		},
	}

	ruleAttrs := map[string]tfschema.Attribute{
		"direction": tfschema.StringAttribute{Computed: true},
		"protocol":  tfschema.StringAttribute{Computed: true},
		"ports": tfschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: portsAttrs,
		},
		"source_refs": tfschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
	}

	resp.Schema = tfschema.Schema{
		Attributes: map[string]tfschema.Attribute{
			"id":                tfschema.StringAttribute{Computed: true},
			"name":              tfschema.StringAttribute{Required: true},
			"workspace_id":      tfschema.StringAttribute{Required: true},
			"tenant":            tfschema.StringAttribute{Computed: true},
			"region":            tfschema.StringAttribute{Computed: true},
			"resource_provider": tfschema.StringAttribute{Computed: true},
			"created_at":        tfschema.StringAttribute{Computed: true},
			"deleted_at":        tfschema.StringAttribute{Computed: true},
			"last_modified_at":  tfschema.StringAttribute{Computed: true},
			"labels": tfschema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"annotations": tfschema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"extensions": tfschema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"rules": tfschema.ListNestedAttribute{
				Computed: true,
				NestedObject: tfschema.NestedAttributeObject{
					Attributes: ruleAttrs,
				},
			},
			"rule_refs": tfschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"state": tfschema.StringAttribute{Computed: true},
		},
	}
}

func (d *SecurityGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(clients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected sdk.Clients, got: %T", req.ProviderData),
		)
		return
	}

	d.client = clients.RegionalClient
	d.tenant = clients.Tenant

	tflog.Debug(ctx, "configured security group data source")
}

func (d *SecurityGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecurityGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "tenant_id", d.tenant)
	ctx = tflog.SetField(ctx, "workspace_id", data.WorkspaceId.ValueString())
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	tflog.Debug(ctx, "reading security group data source")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(d.tenant),
		Workspace: secapi.WorkspaceID(data.WorkspaceId.ValueString()),
		Name:      data.Name.ValueString(),
	}

	sg, err := d.client.NetworkV1.GetSecurityGroup(ctx, wref)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading security group",
			"An error was encountered when reading the security group.\nError: "+err.Error(),
		)
		return
	}

	result, diags := securityGroupToDataSourceModel(ctx, sg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func securityGroupToDataSourceModel(ctx context.Context, sg *sdk.SecurityGroup) (SecurityGroupDataSourceModel, diag.Diagnostics) {
	common, diags := securityGroupToBaseModel(ctx, sg)
	model := SecurityGroupDataSourceModel{securityGroupModel: common}
	if sg.Status != nil {
		model.State = types.StringValue(string(sg.Status.State))
	} else {
		model.State = types.StringNull()
	}
	return model, diags
}
