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
	_ datasource.DataSource              = (*NicDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*NicDataSource)(nil)
)

type NicDataSource struct {
	client *secapi.RegionalClient
	tenant string
}

func newNicDataSource() datasource.DataSource {
	return &NicDataSource{}
}

func (d *NicDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nic"
}

type NicDataSourceModel struct {
	nicModel

	State types.String `tfsdk:"state"`
}

func (d *NicDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"subnet_id": tfschema.StringAttribute{Computed: true},
			"addresses": tfschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"public_ip_ids": tfschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"security_group_ids": tfschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"mac_address": tfschema.StringAttribute{Computed: true},
			"sku_id":      tfschema.StringAttribute{Computed: true},
			"state":       tfschema.StringAttribute{Computed: true},
		},
	}
}

func (d *NicDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured nic data source")
}

func (d *NicDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NicDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "tenant_id", d.tenant)
	ctx = tflog.SetField(ctx, "workspace_id", data.WorkspaceId.ValueString())
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	tflog.Debug(ctx, "reading nic data source")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(d.tenant),
		Workspace: secapi.WorkspaceID(data.WorkspaceId.ValueString()),
		Name:      data.Name.ValueString(),
	}

	nic, err := d.client.NetworkV1.GetNic(ctx, wref)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading nic",
			"An error was encountered when reading the nic.\nError: "+err.Error(),
		)
		return
	}

	result, diags := nicToDataSourceModel(ctx, nic)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func nicToDataSourceModel(ctx context.Context, nic *sdk.Nic) (NicDataSourceModel, diag.Diagnostics) {
	common, diags := nicToBaseModel(ctx, nic)
	model := NicDataSourceModel{nicModel: common}
	if nic.Status != nil {
		model.State = types.StringValue(string(nic.Status.State))
	} else {
		model.State = types.StringNull()
	}
	return model, diags
}
