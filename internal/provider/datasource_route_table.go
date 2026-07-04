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
	_ datasource.DataSource              = (*RouteTableDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*RouteTableDataSource)(nil)
)

type RouteTableDataSource struct {
	client *secapi.RegionalClient
	tenant string
}

func newRouteTableDataSource() datasource.DataSource {
	return &RouteTableDataSource{}
}

func (d *RouteTableDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route_table"
}

type RouteTableDataSourceModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	WorkspaceId      types.String `tfsdk:"workspace_id"`
	NetworkId        types.String `tfsdk:"network_id"`
	Tenant           types.String `tfsdk:"tenant"`
	Region           types.String `tfsdk:"region"`
	ResourceProvider types.String `tfsdk:"resource_provider"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DeletedAt        types.String `tfsdk:"deleted_at"`
	LastModifiedAt   types.String `tfsdk:"last_modified_at"`

	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
	Extensions  types.Map `tfsdk:"extensions"`

	Routes types.List   `tfsdk:"routes"`
	State  types.String `tfsdk:"state"`
}

func (d *RouteTableDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"routes": tfschema.ListNestedAttribute{
				Computed: true,
				NestedObject: tfschema.NestedAttributeObject{
					Attributes: map[string]tfschema.Attribute{
						"destination_cidr_block": tfschema.StringAttribute{Computed: true},
						"target_id":              tfschema.StringAttribute{Computed: true},
					},
				},
			},
			"state": tfschema.StringAttribute{Computed: true},
		},
	}
}

func (d *RouteTableDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured route table data source")
}

func (d *RouteTableDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RouteTableDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "tenant_id", d.tenant)
	ctx = tflog.SetField(ctx, "workspace_id", data.WorkspaceId.ValueString())
	ctx = tflog.SetField(ctx, "network_id", data.NetworkId.ValueString())
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	tflog.Debug(ctx, "reading route table data source")

	nref := secapi.NetworkReference{
		Tenant:    secapi.TenantID(d.tenant),
		Workspace: secapi.WorkspaceID(data.WorkspaceId.ValueString()),
		Network:   secapi.NetworkID(data.NetworkId.ValueString()),
		Name:      data.Name.ValueString(),
	}

	rt, err := d.client.NetworkV1.GetRouteTable(ctx, nref)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading route table",
			"An error was encountered when reading the route table.\nError: "+err.Error(),
		)
		return
	}

	result, diags := routeTableToDataSourceModel(ctx, rt)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func routeTableToDataSourceModel(ctx context.Context, rt *sdk.RouteTable) (RouteTableDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := RouteTableDataSourceModel{}
	model.Id = types.StringValue(rt.Metadata.Ref)
	model.Name = types.StringValue(rt.Metadata.Name)
	model.WorkspaceId = types.StringValue(rt.Metadata.Workspace)
	model.NetworkId = types.StringValue(rt.Metadata.Network)
	model.Tenant = types.StringValue(rt.Metadata.Tenant)
	model.Region = types.StringValue(rt.Metadata.Region)
	model.ResourceProvider = refToResourceProvider(rt.Metadata.Ref)
	model.CreatedAt = fromTime(rt.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(rt.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(rt.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, rt.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, rt.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, rt.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	routes, d := routesToListValue(ctx, rt.Spec.Routes)
	diags.Append(d...)
	model.Routes = routes

	if rt.Status != nil {
		model.State = types.StringValue(string(rt.Status.State))
	} else {
		model.State = types.StringNull()
	}

	return model, diags
}
