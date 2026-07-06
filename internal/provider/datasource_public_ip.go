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
	_ datasource.DataSource              = (*PublicIpDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*PublicIpDataSource)(nil)
)

type PublicIpDataSource struct {
	client *secapi.RegionalClient
	tenant string
}

func newPublicIpDataSource() datasource.DataSource {
	return &PublicIpDataSource{}
}

func (d *PublicIpDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_ip"
}

type PublicIpDataSourceModel struct {
	publicIpModel

	State types.String `tfsdk:"state"`
}

func (d *PublicIpDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"version":     tfschema.StringAttribute{Computed: true},
			"address":     tfschema.StringAttribute{Computed: true},
			"attached_to": tfschema.StringAttribute{Computed: true},
			"ip_address":  tfschema.StringAttribute{Computed: true},
			"state":       tfschema.StringAttribute{Computed: true},
		},
	}
}

func (d *PublicIpDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured public ip data source")
}

func (d *PublicIpDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PublicIpDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "tenant_id", d.tenant)
	ctx = tflog.SetField(ctx, "workspace_id", data.WorkspaceId.ValueString())
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	tflog.Debug(ctx, "reading public ip data source")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(d.tenant),
		Workspace: secapi.WorkspaceID(data.WorkspaceId.ValueString()),
		Name:      data.Name.ValueString(),
	}

	ip, err := d.client.NetworkV1.GetPublicIp(ctx, wref)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading public ip",
			"An error was encountered when reading the public ip.\nError: "+err.Error(),
		)
		return
	}

	result, diags := publicIpToDataSourceModel(ctx, ip)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func publicIpToDataSourceModel(ctx context.Context, ip *sdk.PublicIp) (PublicIpDataSourceModel, diag.Diagnostics) {
	common, diags := publicIpToBaseModel(ctx, ip)
	model := PublicIpDataSourceModel{publicIpModel: common}
	if ip.Status != nil {
		model.State = types.StringValue(string(ip.Status.State))
	} else {
		model.State = types.StringNull()
	}
	return model, diags
}
