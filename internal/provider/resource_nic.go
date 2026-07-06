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
	_ resource.Resource                = (*NicResource)(nil)
	_ resource.ResourceWithConfigure   = (*NicResource)(nil)
	_ resource.ResourceWithImportState = (*NicResource)(nil)
)

type NicResource struct {
	client *secapi.RegionalClient

	tenant string
	region string

	retry retryConfig
}

func newNicResource() resource.Resource {
	return &NicResource{}
}

func (r *NicResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nic"
}

func (r *NicResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

type NicResourceModel struct {
	nicModel

	Retry *RetryModel `tfsdk:"retry"`
}

func (r *NicResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"subnet_id": tfschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"addresses": tfschema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"public_ip_ids": tfschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"security_group_ids": tfschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"mac_address": tfschema.StringAttribute{
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

func (r *NicResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured nic resource")
}

func (r *NicResource) logFields(ctx context.Context, data NicResourceModel) context.Context {
	ctx = tflog.SetField(ctx, "tenant_id", r.tenant)
	ctx = tflog.SetField(ctx, "workspace_id", data.WorkspaceId.ValueString())
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	return ctx
}

func (r *NicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NicResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "creating nic")

	nic, diags := nicFromModel(ctx, r.tenant, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	nic, err := r.client.NetworkV1.CreateOrUpdateNic(ctx, nic)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating nic",
			"An error was encountered when creating the nic.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for nic to become active")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(nic.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(nic.Metadata.Workspace),
		Name:      nic.Metadata.Name,
	}

	nic, err = r.client.NetworkV1.GetNicUntilState(ctx, wref, r.retry.with(data.Retry).untilState(sdk.ResourceStateActive))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading nic",
			"An error was encountered while waiting for the nic to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags := nicToResourceModel(ctx, nic)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "nic created")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *NicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "reading nic")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(r.tenant),
		Workspace: secapi.WorkspaceID(data.WorkspaceId.ValueString()),
		Name:      data.Name.ValueString(),
	}

	nic, err := r.client.NetworkV1.GetNic(ctx, wref)
	if err == secapi.ErrResourceNotFound {
		tflog.Debug(ctx, "nic not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error reading nic",
			"An error was encountered when reading the nic.\nError: "+err.Error(),
		)
		return
	}

	result, diags := nicToResourceModel(ctx, nic)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *NicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NicResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "updating nic")

	nic, diags := nicFromModel(ctx, r.tenant, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	nic, err := r.client.NetworkV1.CreateOrUpdateNic(ctx, nic)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating nic",
			"An error was encountered when updating the nic.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for nic to become active")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(nic.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(nic.Metadata.Workspace),
		Name:      nic.Metadata.Name,
	}

	nic, err = r.client.NetworkV1.GetNicUntilState(ctx, wref, r.retry.with(data.Retry).untilState(sdk.ResourceStateActive))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading nic",
			"An error was encountered while waiting for the nic to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags := nicToResourceModel(ctx, nic)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "nic updated")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *NicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "deleting nic")

	nic := &sdk.Nic{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Tenant:    r.tenant,
			Workspace: data.WorkspaceId.ValueString(),
			Name:      data.Name.ValueString(),
		},
	}

	err := r.client.NetworkV1.DeleteNic(ctx, nic)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting nic",
			"An error was encountered when deleting the nic.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for nic to be deleted")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(nic.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(nic.Metadata.Workspace),
		Name:      nic.Metadata.Name,
	}

	err = r.client.NetworkV1.WatchNicUntilDeleted(ctx, wref, r.retry.with(data.Retry).observer())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading nic",
			"An error was encountered while waiting for the nic to become deleted.\nError: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "nic deleted")
}

func nicFromModel(ctx context.Context, tenant string, data NicResourceModel) (*sdk.Nic, diag.Diagnostics) {
	var diags diag.Diagnostics

	var addresses []string
	if !data.Addresses.IsNull() && !data.Addresses.IsUnknown() {
		diags.Append(data.Addresses.ElementsAs(ctx, &addresses, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}

	nic := &sdk.Nic{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Tenant:    tenant,
			Workspace: data.WorkspaceId.ValueString(),
			Name:      data.Name.ValueString(),
		},
		Labels:      toStringMap(data.Labels),
		Annotations: toStringMap(data.Annotations),
		Extensions:  toStringMap(data.Extensions),
		Spec: sdk.NicSpec{
			SubnetRef: sdk.Reference{
				Resource: data.SubnetId.ValueString(),
			},
			Addresses: addresses,
		},
	}

	return nic, diags
}

func nicToResourceModel(ctx context.Context, nic *sdk.Nic) (NicResourceModel, diag.Diagnostics) {
	common, diags := nicToBaseModel(ctx, nic)
	return NicResourceModel{nicModel: common}, diags
}

func refsToStringList(strs []string) (types.List, diag.Diagnostics) {
	if len(strs) == 0 {
		return types.ListValueMust(types.StringType, []attr.Value{}), nil
	}
	elems := make([]attr.Value, 0, len(strs))
	for _, s := range strs {
		elems = append(elems, types.StringValue(s))
	}
	return types.ListValue(types.StringType, elems)
}

func refsToStringListFromRefs(refs []sdk.Reference) (types.List, diag.Diagnostics) {
	if len(refs) == 0 {
		return types.ListValueMust(types.StringType, []attr.Value{}), nil
	}
	elems := make([]attr.Value, 0, len(refs))
	for _, ref := range refs {
		elems = append(elems, types.StringValue(ref.Resource))
	}
	return types.ListValue(types.StringType, elems)
}
