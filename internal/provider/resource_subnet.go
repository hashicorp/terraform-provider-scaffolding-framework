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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
	"github.com/eu-sovereign-cloud/go-sdk/secapi"
)

var (
	_ resource.Resource                = (*SubnetResource)(nil)
	_ resource.ResourceWithConfigure   = (*SubnetResource)(nil)
	_ resource.ResourceWithImportState = (*SubnetResource)(nil)
)

var subnetCidrAttrTypes = map[string]attr.Type{
	"ipv4": types.StringType,
	"ipv6": types.StringType,
}

type SubnetResource struct {
	client *secapi.RegionalClient

	tenant string
	region string

	retry retryConfig
}

func newSubnetResource() resource.Resource {
	return &SubnetResource{}
}

func (r *SubnetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (r *SubnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

type SubnetCidrModel struct {
	Ipv4 types.String `tfsdk:"ipv4"`
	Ipv6 types.String `tfsdk:"ipv6"`
}

type SubnetModel struct {
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

	Cidr         SubnetCidrModel `tfsdk:"cidr"`
	RouteTableId types.String    `tfsdk:"route_table_id"`
	Zone         types.String    `tfsdk:"zone"`
	SkuId        types.String    `tfsdk:"sku_id"`

	Retry *RetryModel `tfsdk:"retry"`
}

func (r *SubnetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	cidrAttrs := map[string]tfschema.Attribute{
		"ipv4": tfschema.StringAttribute{Optional: true, Computed: true},
		"ipv6": tfschema.StringAttribute{Optional: true, Computed: true},
	}

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
			"cidr": tfschema.SingleNestedAttribute{
				Required:   true,
				Attributes: cidrAttrs,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
			"route_table_id": tfschema.StringAttribute{
				Optional: true,
			},
			"zone": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sku_id": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"retry": retryResourceSchema(),
		},
	}
}

func (r *SubnetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured subnet resource")
}

func (r *SubnetResource) logFields(ctx context.Context, data SubnetModel) context.Context {
	ctx = tflog.SetField(ctx, "tenant_id", r.tenant)
	ctx = tflog.SetField(ctx, "workspace_id", data.WorkspaceId.ValueString())
	ctx = tflog.SetField(ctx, "network_id", data.NetworkId.ValueString())
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	return ctx
}

func (r *SubnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubnetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "creating subnet")

	sub := subnetFromModel(r.tenant, data)

	sub, err := r.client.NetworkV1.CreateOrUpdateSubnet(ctx, sub)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating subnet",
			"An error was encountered when creating the subnet.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for subnet to become active")

	nref := secapi.NetworkReference{
		Tenant:    secapi.TenantID(sub.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(sub.Metadata.Workspace),
		Network:   secapi.NetworkID(sub.Metadata.Network),
		Name:      sub.Metadata.Name,
	}

	sub, err = r.client.NetworkV1.GetSubnetUntilState(ctx, nref, r.retry.with(data.Retry).untilState(sdk.ResourceStateActive))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading subnet",
			"An error was encountered while waiting for the subnet to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags := subnetToResourceModel(ctx, sub)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "subnet created")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *SubnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubnetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "reading subnet")

	nref := secapi.NetworkReference{
		Tenant:    secapi.TenantID(r.tenant),
		Workspace: secapi.WorkspaceID(data.WorkspaceId.ValueString()),
		Network:   secapi.NetworkID(data.NetworkId.ValueString()),
		Name:      data.Name.ValueString(),
	}

	sub, err := r.client.NetworkV1.GetSubnet(ctx, nref)
	if err == secapi.ErrResourceNotFound {
		tflog.Debug(ctx, "subnet not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error reading subnet",
			"An error was encountered when reading the subnet.\nError: "+err.Error(),
		)
		return
	}

	result, diags := subnetToResourceModel(ctx, sub)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *SubnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SubnetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "updating subnet")

	sub := subnetFromModel(r.tenant, data)

	sub, err := r.client.NetworkV1.CreateOrUpdateSubnet(ctx, sub)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating subnet",
			"An error was encountered when updating the subnet.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for subnet to become active")

	nref := secapi.NetworkReference{
		Tenant:    secapi.TenantID(sub.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(sub.Metadata.Workspace),
		Network:   secapi.NetworkID(sub.Metadata.Network),
		Name:      sub.Metadata.Name,
	}

	sub, err = r.client.NetworkV1.GetSubnetUntilState(ctx, nref, r.retry.with(data.Retry).untilState(sdk.ResourceStateActive))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading subnet",
			"An error was encountered while waiting for the subnet to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags := subnetToResourceModel(ctx, sub)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "subnet updated")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *SubnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubnetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "deleting subnet")

	sub := &sdk.Subnet{
		Metadata: &sdk.RegionalNetworkResourceMetadata{
			Tenant:    r.tenant,
			Workspace: data.WorkspaceId.ValueString(),
			Network:   data.NetworkId.ValueString(),
			Name:      data.Name.ValueString(),
		},
	}

	err := r.client.NetworkV1.DeleteSubnet(ctx, sub)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting subnet",
			"An error was encountered when deleting the subnet.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for subnet to be deleted")

	nref := secapi.NetworkReference{
		Tenant:    secapi.TenantID(sub.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(sub.Metadata.Workspace),
		Network:   secapi.NetworkID(sub.Metadata.Network),
		Name:      sub.Metadata.Name,
	}

	err = r.client.NetworkV1.WatchSubnetUntilDeleted(ctx, nref, r.retry.with(data.Retry).observer())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading subnet",
			"An error was encountered while waiting for the subnet to become deleted.\nError: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "subnet deleted")
}

func subnetFromModel(tenant string, data SubnetModel) *sdk.Subnet {
	sub := &sdk.Subnet{
		Metadata: &sdk.RegionalNetworkResourceMetadata{
			Tenant:    tenant,
			Workspace: data.WorkspaceId.ValueString(),
			Network:   data.NetworkId.ValueString(),
			Name:      data.Name.ValueString(),
		},
		Labels:      toStringMap(data.Labels),
		Annotations: toStringMap(data.Annotations),
		Extensions:  toStringMap(data.Extensions),
		Spec: sdk.SubnetSpec{
			Cidr: sdk.Cidr{
				Ipv4: data.Cidr.Ipv4.ValueString(),
				Ipv6: data.Cidr.Ipv6.ValueString(),
			},
		},
	}

	if !data.RouteTableId.IsNull() && !data.RouteTableId.IsUnknown() {
		sub.Spec.RouteTableRef = sdk.Reference{
			Resource: data.RouteTableId.ValueString(),
		}
	}

	return sub
}

func subnetToResourceModel(ctx context.Context, sub *sdk.Subnet) (SubnetModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := SubnetModel{}
	model.Id = types.StringValue(sub.Metadata.Ref)
	model.Name = types.StringValue(sub.Metadata.Name)
	model.WorkspaceId = types.StringValue(sub.Metadata.Workspace)
	model.NetworkId = types.StringValue(sub.Metadata.Network)
	model.Tenant = types.StringValue(sub.Metadata.Tenant)
	model.Region = types.StringValue(sub.Metadata.Region)
	model.ResourceProvider = refToResourceProvider(sub.Metadata.Ref)
	model.CreatedAt = fromTime(sub.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(sub.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(sub.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, sub.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, sub.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, sub.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	model.Cidr = SubnetCidrModel{
		Ipv4: types.StringValue(sub.Spec.Cidr.Ipv4),
		Ipv6: types.StringValue(sub.Spec.Cidr.Ipv6),
	}

	model.RouteTableId = types.StringValue(sub.Spec.RouteTableRef.Resource)
	model.Zone = types.StringValue(sub.Spec.Zone)
	model.SkuId = fromRefPtr(sub.Spec.SkuRef)

	return model, diags
}
