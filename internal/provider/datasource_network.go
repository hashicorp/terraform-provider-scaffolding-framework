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
	_ datasource.DataSource              = (*NetworkDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*NetworkDataSource)(nil)
)

type NetworkDataSource struct {
	client *secapi.RegionalClient
	tenant string
}

func newNetworkDataSource() datasource.DataSource {
	return &NetworkDataSource{}
}

func (d *NetworkDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

type NetworkDataSourceModel struct {
	networkModel

	State types.String `tfsdk:"state"`
}

func (d *NetworkDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	cidrAttrs := map[string]tfschema.Attribute{
		"ipv4": tfschema.StringAttribute{Computed: true},
		"ipv6": tfschema.StringAttribute{Computed: true},
	}

	resp.Schema = tfschema.Schema{
		Attributes: map[string]tfschema.Attribute{
			"id": tfschema.StringAttribute{Computed: true},
			"name": tfschema.StringAttribute{
				Required: true,
			},
			"workspace_id": tfschema.StringAttribute{
				Required: true,
			},
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
			"sku_id": tfschema.StringAttribute{Computed: true},
			"cidr": tfschema.SingleNestedAttribute{
				Computed:   true,
				Attributes: cidrAttrs,
			},
			"additional_cidrs": tfschema.ListNestedAttribute{
				Computed: true,
				NestedObject: tfschema.NestedAttributeObject{
					Attributes: cidrAttrs,
				},
			},
			"state": tfschema.StringAttribute{Computed: true},
		},
	}
}

func (d *NetworkDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured network data source")
}

func (d *NetworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NetworkDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "tenant_id", d.tenant)
	ctx = tflog.SetField(ctx, "workspace_id", data.WorkspaceId.ValueString())
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	tflog.Debug(ctx, "reading network data source")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(d.tenant),
		Workspace: secapi.WorkspaceID(data.WorkspaceId.ValueString()),
		Name:      data.Name.ValueString(),
	}

	net, err := d.client.NetworkV1.GetNetwork(ctx, wref)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading network",
			"An error was encountered when reading the network.\nError: "+err.Error(),
		)
		return
	}

	data, diags := networkToDataSourceModel(ctx, net)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func networkToDataSourceModel(ctx context.Context, net *sdk.Network) (NetworkDataSourceModel, diag.Diagnostics) {
	common, diags := networkToBaseModel(ctx, net)
	model := NetworkDataSourceModel{networkModel: common}
	if net.Status != nil {
		model.State = types.StringValue(string(net.Status.State))
	} else {
		model.State = types.StringNull()
	}
	return model, diags
}
