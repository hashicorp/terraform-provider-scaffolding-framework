package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	tfschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
	"github.com/eu-sovereign-cloud/go-sdk/secapi"
)

var (
	_ resource.Resource                = (*RouteTableResource)(nil)
	_ resource.ResourceWithConfigure   = (*RouteTableResource)(nil)
	_ resource.ResourceWithImportState = (*RouteTableResource)(nil)
)

var routeAttrTypes = map[string]attr.Type{
	"destination_cidr_block": types.StringType,
	"target_id":              types.StringType,
}

type RouteTableResource struct {
	client *secapi.RegionalClient

	tenant string
	region string

	retry retryConfig
}

func newRouteTableResource() resource.Resource {
	return &RouteTableResource{}
}

func (r *RouteTableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route_table"
}

func (r *RouteTableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier in the format \"workspace_id/network_id/name\", got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("network_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[2])...)
}

type RouteModel struct {
	DestinationCidrBlock types.String `tfsdk:"destination_cidr_block"`
	TargetId             types.String `tfsdk:"target_id"`
}

type RouteTableResourceModel struct {
	routeTableModel

	Retry *RetryModel `tfsdk:"retry"`
}

func (r *RouteTableResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Attributes: map[string]tfschema.Attribute{
			"id": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": tfschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"workspace_id": tfschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_id": tfschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tenant": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_provider": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deleted_at": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_modified_at": tfschema.StringAttribute{
				Computed: true,
			},
			"labels": tfschema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"annotations": tfschema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"extensions": tfschema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"routes": tfschema.ListNestedAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: tfschema.NestedAttributeObject{
					Attributes: map[string]tfschema.Attribute{
						"destination_cidr_block": tfschema.StringAttribute{Required: true},
						"target_id":              tfschema.StringAttribute{Required: true},
					},
				},
			},
			"retry": retryResourceSchema(),
		},
	}
}

func (r *RouteTableResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = clients.RegionalClient
	r.tenant = clients.Tenant
	r.region = clients.Region
	r.retry = retryConfig{
		delay:       clients.RetryDelay,
		interval:    clients.RetryInterval,
		maxAttempts: clients.RetryMaxAttempts,
	}

	tflog.Debug(ctx, "configured route table resource")
}

func (r *RouteTableResource) logFields(ctx context.Context, data RouteTableResourceModel) context.Context {
	ctx = tflog.SetField(ctx, "tenant_id", r.tenant)
	ctx = tflog.SetField(ctx, "workspace_id", data.WorkspaceId.ValueString())
	ctx = tflog.SetField(ctx, "network_id", data.NetworkId.ValueString())
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	return ctx
}

func (r *RouteTableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RouteTableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "creating route table")

	rt, diags := routeTableFromModel(ctx, r.tenant, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rt, err := r.client.NetworkV1.CreateOrUpdateRouteTable(ctx, rt)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating route table",
			"An error was encountered when creating the route table.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for route table to become active")

	nref := secapi.NetworkReference{
		Tenant:    secapi.TenantID(rt.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(rt.Metadata.Workspace),
		Network:   secapi.NetworkID(rt.Metadata.Network),
		Name:      rt.Metadata.Name,
	}

	rt, err = r.client.NetworkV1.GetRouteTableUntilState(ctx, nref, r.retry.with(data.Retry).untilState(sdk.ResourceStateActive))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading route table",
			"An error was encountered while waiting for the route table to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags := routeTableToResourceModel(ctx, rt)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "route table created")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *RouteTableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RouteTableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "reading route table")

	nref := secapi.NetworkReference{
		Tenant:    secapi.TenantID(r.tenant),
		Workspace: secapi.WorkspaceID(data.WorkspaceId.ValueString()),
		Network:   secapi.NetworkID(data.NetworkId.ValueString()),
		Name:      data.Name.ValueString(),
	}

	rt, err := r.client.NetworkV1.GetRouteTable(ctx, nref)
	if err == secapi.ErrResourceNotFound {
		tflog.Debug(ctx, "route table not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error reading route table",
			"An error was encountered when reading the route table.\nError: "+err.Error(),
		)
		return
	}

	result, diags := routeTableToResourceModel(ctx, rt)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *RouteTableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RouteTableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "updating route table")

	rt, diags := routeTableFromModel(ctx, r.tenant, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rt, err := r.client.NetworkV1.CreateOrUpdateRouteTable(ctx, rt)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating route table",
			"An error was encountered when updating the route table.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for route table to become active")

	nref := secapi.NetworkReference{
		Tenant:    secapi.TenantID(rt.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(rt.Metadata.Workspace),
		Network:   secapi.NetworkID(rt.Metadata.Network),
		Name:      rt.Metadata.Name,
	}

	rt, err = r.client.NetworkV1.GetRouteTableUntilState(ctx, nref, r.retry.with(data.Retry).untilState(sdk.ResourceStateActive))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading route table",
			"An error was encountered while waiting for the route table to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags := routeTableToResourceModel(ctx, rt)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "route table updated")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *RouteTableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RouteTableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "deleting route table")

	rt := &sdk.RouteTable{
		Metadata: &sdk.RegionalNetworkResourceMetadata{
			Tenant:    r.tenant,
			Workspace: data.WorkspaceId.ValueString(),
			Network:   data.NetworkId.ValueString(),
			Name:      data.Name.ValueString(),
		},
	}

	err := r.client.NetworkV1.DeleteRouteTable(ctx, rt)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting route table",
			"An error was encountered when deleting the route table.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for route table to be deleted")

	nref := secapi.NetworkReference{
		Tenant:    secapi.TenantID(rt.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(rt.Metadata.Workspace),
		Network:   secapi.NetworkID(rt.Metadata.Network),
		Name:      rt.Metadata.Name,
	}

	err = r.client.NetworkV1.WatchRouteTableUntilDeleted(ctx, nref, r.retry.with(data.Retry).observer())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading route table",
			"An error was encountered while waiting for the route table to become deleted.\nError: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "route table deleted")
}

func routeTableFromModel(ctx context.Context, tenant string, data RouteTableResourceModel) (*sdk.RouteTable, diag.Diagnostics) {
	var diags diag.Diagnostics

	routes, d := routesFromModel(ctx, data.Routes)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	rt := &sdk.RouteTable{
		Metadata: &sdk.RegionalNetworkResourceMetadata{
			Tenant:    tenant,
			Workspace: data.WorkspaceId.ValueString(),
			Network:   data.NetworkId.ValueString(),
			Name:      data.Name.ValueString(),
		},
		Labels:      toStringMap(data.Labels),
		Annotations: toStringMap(data.Annotations),
		Extensions:  toStringMap(data.Extensions),
		Spec: sdk.RouteTableSpec{
			Routes: routes,
		},
	}

	return rt, diags
}

func routesFromModel(ctx context.Context, list types.List) ([]sdk.RouteSpec, diag.Diagnostics) {
	var diags diag.Diagnostics

	if list.IsNull() || list.IsUnknown() {
		return nil, diags
	}

	var routes []RouteModel
	diags.Append(list.ElementsAs(ctx, &routes, false)...)
	if diags.HasError() {
		return nil, diags
	}

	specs := make([]sdk.RouteSpec, 0, len(routes))
	for _, r := range routes {
		specs = append(specs, sdk.RouteSpec{
			DestinationCidrBlock: r.DestinationCidrBlock.ValueString(),
			TargetRef: sdk.Reference{
				Resource: r.TargetId.ValueString(),
			},
		})
	}

	return specs, diags
}

func routeTableToResourceModel(ctx context.Context, rt *sdk.RouteTable) (RouteTableResourceModel, diag.Diagnostics) {
	common, diags := routeTableToBaseModel(ctx, rt)
	return RouteTableResourceModel{routeTableModel: common}, diags
}

func routesToListValue(_ context.Context, specs []sdk.RouteSpec) (types.List, diag.Diagnostics) {
	objType := types.ObjectType{AttrTypes: routeAttrTypes}

	if len(specs) == 0 {
		return types.ListValueMust(objType, []attr.Value{}), nil
	}

	elems := make([]attr.Value, 0, len(specs))
	for _, s := range specs {
		obj, diags := types.ObjectValue(routeAttrTypes, map[string]attr.Value{
			"destination_cidr_block": types.StringValue(s.DestinationCidrBlock),
			"target_id":              types.StringValue(s.TargetRef.Resource),
		})
		if diags.HasError() {
			return types.ListNull(objType), diags
		}
		elems = append(elems, obj)
	}

	return types.ListValue(objType, elems)
}
