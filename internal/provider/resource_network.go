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
	_ resource.Resource                = (*NetworkResource)(nil)
	_ resource.ResourceWithConfigure   = (*NetworkResource)(nil)
	_ resource.ResourceWithImportState = (*NetworkResource)(nil)
)

type NetworkResource struct {
	client *secapi.RegionalClient

	tenant string
	region string

	retry retryConfig
}

func newNetworkResource() resource.Resource {
	return &NetworkResource{}
}

func (r *NetworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	workspaceID, name, ok := strings.Cut(req.ID, "/")
	if !ok || workspaceID == "" || name == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier in the format \"workspace_id/name\", got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), workspaceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
}

// NetworkCidrModel represents an IPv4/IPv6 CIDR block.
type NetworkCidrModel struct {
	Ipv4 types.String `tfsdk:"ipv4"`
	Ipv6 types.String `tfsdk:"ipv6"`
}

var networkCidrAttrTypes = map[string]attr.Type{
	"ipv4": types.StringType,
	"ipv6": types.StringType,
}

type NetworkModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	WorkspaceId      types.String `tfsdk:"workspace_id"`
	Tenant           types.String `tfsdk:"tenant"`
	Region           types.String `tfsdk:"region"`
	ResourceProvider types.String `tfsdk:"resource_provider"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DeletedAt        types.String `tfsdk:"deleted_at"`
	LastModifiedAt   types.String `tfsdk:"last_modified_at"`

	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
	Extensions  types.Map `tfsdk:"extensions"`

	SkuId           types.String     `tfsdk:"sku_id"`
	Cidr            NetworkCidrModel `tfsdk:"cidr"`
	AdditionalCidrs types.List       `tfsdk:"additional_cidrs"`

	Retry *RetryModel `tfsdk:"retry"`
}

func (r *NetworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	cidrAttrs := map[string]tfschema.Attribute{
		"ipv4": tfschema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"ipv6": tfschema.StringAttribute{
			Optional: true,
			Computed: true,
		},
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
			// sku_id is immutable after creation per SDK spec
			"sku_id": tfschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// cidr is immutable after creation per SDK spec
			"cidr": tfschema.SingleNestedAttribute{
				Required:   true,
				Attributes: cidrAttrs,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
			"additional_cidrs": tfschema.ListNestedAttribute{
				Optional: true,
				NestedObject: tfschema.NestedAttributeObject{
					Attributes: cidrAttrs,
				},
			},
			"retry": retryResourceSchema(),
		},
	}
}

func (r *NetworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured network resource")
}

func (r *NetworkResource) logFields(ctx context.Context, data NetworkModel) context.Context {
	ctx = tflog.SetField(ctx, "tenant_id", r.tenant)
	ctx = tflog.SetField(ctx, "workspace_id", data.WorkspaceId.ValueString())
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	return ctx
}

func (r *NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NetworkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "creating network")

	net := networkFromModel(ctx, r.tenant, data)

	net, err := r.client.NetworkV1.CreateOrUpdateNetwork(ctx, net)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating network",
			"An error was encountered when creating the network.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for network to become active")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(net.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(net.Metadata.Workspace),
		Name:      net.Metadata.Name,
	}

	net, err = r.client.NetworkV1.GetNetworkUntilState(ctx, wref, r.retry.with(data.Retry).untilState(sdk.ResourceStateActive))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading network",
			"An error was encountered while waiting for the network to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags := networkToResourceModel(ctx, net)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "network created")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *NetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NetworkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "reading network")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(r.tenant),
		Workspace: secapi.WorkspaceID(data.WorkspaceId.ValueString()),
		Name:      data.Name.ValueString(),
	}

	net, err := r.client.NetworkV1.GetNetwork(ctx, wref)
	if err == secapi.ErrResourceNotFound {
		tflog.Debug(ctx, "network not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error reading network",
			"An error was encountered when reading the network.\nError: "+err.Error(),
		)
		return
	}

	result, diags := networkToResourceModel(ctx, net)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NetworkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "updating network")

	net := networkFromModel(ctx, r.tenant, data)

	net, err := r.client.NetworkV1.CreateOrUpdateNetwork(ctx, net)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating network",
			"An error was encountered when updating the network.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for network to become active")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(net.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(net.Metadata.Workspace),
		Name:      net.Metadata.Name,
	}

	net, err = r.client.NetworkV1.GetNetworkUntilState(ctx, wref, r.retry.with(data.Retry).untilState(sdk.ResourceStateActive))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading network",
			"An error was encountered while waiting for the network to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags := networkToResourceModel(ctx, net)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "network updated")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *NetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NetworkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "deleting network")

	net := &sdk.Network{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Tenant:    r.tenant,
			Workspace: data.WorkspaceId.ValueString(),
			Name:      data.Name.ValueString(),
		},
	}

	err := r.client.NetworkV1.DeleteNetwork(ctx, net)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting network",
			"An error was encountered when deleting the network.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for network to be deleted")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(net.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(net.Metadata.Workspace),
		Name:      net.Metadata.Name,
	}

	err = r.client.NetworkV1.WatchNetworkUntilDeleted(ctx, wref, r.retry.with(data.Retry).observer())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading network",
			"An error was encountered while waiting for the network to become deleted.\nError: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "network deleted")
}

func networkFromModel(ctx context.Context, tenant string, data NetworkModel) *sdk.Network {
	net := &sdk.Network{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Tenant:    tenant,
			Workspace: data.WorkspaceId.ValueString(),
			Name:      data.Name.ValueString(),
		},
		Labels:      toStringMap(data.Labels),
		Annotations: toStringMap(data.Annotations),
		Extensions:  toStringMap(data.Extensions),
		Spec: sdk.NetworkSpec{
			SkuRef: sdk.Reference{
				Resource: data.SkuId.ValueString(),
			},
			Cidr: sdk.Cidr{
				Ipv4: data.Cidr.Ipv4.ValueString(),
				Ipv6: data.Cidr.Ipv6.ValueString(),
			},
		},
	}

	if !data.AdditionalCidrs.IsNull() && !data.AdditionalCidrs.IsUnknown() {
		var cidrs []NetworkCidrModel
		data.AdditionalCidrs.ElementsAs(ctx, &cidrs, false)
		for _, c := range cidrs {
			net.Spec.AdditionalCidrs = append(net.Spec.AdditionalCidrs, sdk.Cidr{
				Ipv4: c.Ipv4.ValueString(),
				Ipv6: c.Ipv6.ValueString(),
			})
		}
	}

	return net
}

func networkToResourceModel(ctx context.Context, net *sdk.Network) (NetworkModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := NetworkModel{}
	model.Id = types.StringValue(net.Metadata.Ref)
	model.Name = types.StringValue(net.Metadata.Name)
	model.WorkspaceId = types.StringValue(net.Metadata.Workspace)
	model.Tenant = types.StringValue(net.Metadata.Tenant)
	model.Region = types.StringValue(net.Metadata.Region)
	model.ResourceProvider = refToResourceProvider(net.Metadata.Ref)
	model.CreatedAt = fromTime(net.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(net.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(net.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, net.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, net.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, net.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	model.SkuId = types.StringValue(net.Spec.SkuRef.Resource)

	// Read CIDR from status (authoritative resolved values after provisioning).
	// Fall back to spec if status is not yet populated.
	if net.Status != nil {
		model.Cidr = cidrFromSDK(net.Status.Cidr)
		additionalCidrs, d := fromCidrList(ctx, net.Status.AdditionalCidrs)
		diags.Append(d...)
		model.AdditionalCidrs = additionalCidrs
	} else {
		model.Cidr = cidrFromSDK(net.Spec.Cidr)
		additionalCidrs, d := fromCidrList(ctx, net.Spec.AdditionalCidrs)
		diags.Append(d...)
		model.AdditionalCidrs = additionalCidrs
	}

	return model, diags
}

func cidrFromSDK(c sdk.Cidr) NetworkCidrModel {
	return NetworkCidrModel{
		Ipv4: fromNonEmptyString(c.Ipv4),
		Ipv6: fromNonEmptyString(c.Ipv6),
	}
}

func fromCidrList(ctx context.Context, cidrs []sdk.Cidr) (types.List, diag.Diagnostics) {
	if len(cidrs) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: networkCidrAttrTypes}), nil
	}
	models := make([]NetworkCidrModel, 0, len(cidrs))
	for _, c := range cidrs {
		models = append(models, cidrFromSDK(c))
	}
	return types.ListValueFrom(ctx, types.ObjectType{AttrTypes: networkCidrAttrTypes}, models)
}
