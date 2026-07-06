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
	_ datasource.DataSource              = (*SubnetDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*SubnetDataSource)(nil)
)

type SubnetDataSource struct {
	client *secapi.RegionalClient
	tenant string
}

func newSubnetDataSource() datasource.DataSource {
	return &SubnetDataSource{}
}

func (d *SubnetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

type SubnetDataSourceModel struct {
	subnetModel

	State types.String `tfsdk:"state"`
}

func (d *SubnetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Attributes: map[string]tfschema.Attribute{
			"id":                tfschema.StringAttribute{Computed: true},
			"name":              tfschema.StringAttribute{Required: true},
			"workspace_id":      tfschema.StringAttribute{Required: true},
			"network_id":        tfschema.StringAttribute{Required: true},
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
			"cidr": tfschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]tfschema.Attribute{
					"ipv4": tfschema.StringAttribute{Computed: true},
					"ipv6": tfschema.StringAttribute{Computed: true},
				},
			},
			"route_table_id": tfschema.StringAttribute{Computed: true},
			"zone":           tfschema.StringAttribute{Computed: true},
			"sku_id":         tfschema.StringAttribute{Computed: true},
			"state":          tfschema.StringAttribute{Computed: true},
		},
	}
}

func (d *SubnetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured subnet data source")
}

func (d *SubnetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubnetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "tenant_id", d.tenant)
	ctx = tflog.SetField(ctx, "workspace_id", data.WorkspaceId.ValueString())
	ctx = tflog.SetField(ctx, "network_id", data.NetworkId.ValueString())
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	tflog.Debug(ctx, "reading subnet data source")

	nref := secapi.NetworkReference{
		Tenant:    secapi.TenantID(d.tenant),
		Workspace: secapi.WorkspaceID(data.WorkspaceId.ValueString()),
		Network:   secapi.NetworkID(data.NetworkId.ValueString()),
		Name:      data.Name.ValueString(),
	}

	sub, err := d.client.NetworkV1.GetSubnet(ctx, nref)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading subnet",
			"An error was encountered when reading the subnet.\nError: "+err.Error(),
		)
		return
	}

	result, diags := subnetToDataSourceModel(ctx, sub)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func subnetToDataSourceModel(ctx context.Context, sub *sdk.Subnet) (SubnetDataSourceModel, diag.Diagnostics) {
	common, diags := subnetToBaseModel(ctx, sub)
	model := SubnetDataSourceModel{subnetModel: common}
	if sub.Status != nil {
		model.State = types.StringValue(string(sub.Status.State))
	} else {
		model.State = types.StringNull()
	}
	return model, diags
}
