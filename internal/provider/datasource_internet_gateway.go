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
	_ datasource.DataSource              = (*InternetGatewayDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*InternetGatewayDataSource)(nil)
)

type InternetGatewayDataSource struct {
	client *secapi.RegionalClient
	tenant string
}

func newInternetGatewayDataSource() datasource.DataSource {
	return &InternetGatewayDataSource{}
}

func (d *InternetGatewayDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_internet_gateway"
}

type InternetGatewayDataSourceModel struct {
	internetGatewayModel

	State types.String `tfsdk:"state"`
}

func (d *InternetGatewayDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"egress_only": tfschema.BoolAttribute{Computed: true},
			"state":       tfschema.StringAttribute{Computed: true},
		},
	}
}

func (d *InternetGatewayDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured internet gateway data source")
}

func (d *InternetGatewayDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InternetGatewayDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "tenant_id", d.tenant)
	ctx = tflog.SetField(ctx, "workspace_id", data.WorkspaceId.ValueString())
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	tflog.Debug(ctx, "reading internet gateway data source")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(d.tenant),
		Workspace: secapi.WorkspaceID(data.WorkspaceId.ValueString()),
		Name:      data.Name.ValueString(),
	}

	gtw, err := d.client.NetworkV1.GetInternetGateway(ctx, wref)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading internet gateway",
			"An error was encountered when reading the internet gateway.\nError: "+err.Error(),
		)
		return
	}

	data, diags := internetGatewayToDataSourceModel(ctx, gtw)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func internetGatewayToDataSourceModel(ctx context.Context, gtw *sdk.InternetGateway) (InternetGatewayDataSourceModel, diag.Diagnostics) {
	common, diags := internetGatewayToBaseModel(ctx, gtw)
	model := InternetGatewayDataSourceModel{internetGatewayModel: common}
	if gtw.Status != nil {
		model.State = types.StringValue(string(gtw.Status.State))
	} else {
		model.State = types.StringNull()
	}
	return model, diags
}
