package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	tfschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
	"github.com/eu-sovereign-cloud/go-sdk/secapi"
)

var (
	_ resource.Resource                = (*PublicIpResource)(nil)
	_ resource.ResourceWithConfigure   = (*PublicIpResource)(nil)
	_ resource.ResourceWithImportState = (*PublicIpResource)(nil)
)

type PublicIpResource struct {
	client *secapi.RegionalClient

	tenant string
	region string

	retry retryConfig
}

func newPublicIpResource() resource.Resource {
	return &PublicIpResource{}
}

func (r *PublicIpResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_ip"
}

func (r *PublicIpResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

type PublicIpResourceModel struct {
	publicIpModel

	Retry *RetryModel `tfsdk:"retry"`
}

func (r *PublicIpResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"version": tfschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"address": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"attached_to": tfschema.StringAttribute{
				Computed: true,
			},
			"ip_address": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"retry": retryResourceSchema(),
		},
	}
}

func (r *PublicIpResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured public ip resource")
}

func (r *PublicIpResource) logFields(ctx context.Context, data PublicIpResourceModel) context.Context {
	ctx = tflog.SetField(ctx, "tenant_id", r.tenant)
	ctx = tflog.SetField(ctx, "workspace_id", data.WorkspaceId.ValueString())
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	return ctx
}

func (r *PublicIpResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PublicIpResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "creating public ip")

	ip := publicIpFromModel(r.tenant, data)

	ip, err := r.client.NetworkV1.CreateOrUpdatePublicIp(ctx, ip)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating public ip",
			"An error was encountered when creating the public ip.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for public ip to become active")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(ip.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(ip.Metadata.Workspace),
		Name:      ip.Metadata.Name,
	}

	ip, err = r.client.NetworkV1.GetPublicIpUntilState(ctx, wref, r.retry.with(data.Retry).untilState(sdk.ResourceStateActive))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading public ip",
			"An error was encountered while waiting for the public ip to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags := publicIpToResourceModel(ctx, ip)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "public ip created")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *PublicIpResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PublicIpResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "reading public ip")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(r.tenant),
		Workspace: secapi.WorkspaceID(data.WorkspaceId.ValueString()),
		Name:      data.Name.ValueString(),
	}

	ip, err := r.client.NetworkV1.GetPublicIp(ctx, wref)
	if err == secapi.ErrResourceNotFound {
		tflog.Debug(ctx, "public ip not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error reading public ip",
			"An error was encountered when reading the public ip.\nError: "+err.Error(),
		)
		return
	}

	result, diags := publicIpToResourceModel(ctx, ip)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *PublicIpResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PublicIpResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "updating public ip")

	ip := publicIpFromModel(r.tenant, data)

	ip, err := r.client.NetworkV1.CreateOrUpdatePublicIp(ctx, ip)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating public ip",
			"An error was encountered when updating the public ip.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for public ip to become active")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(ip.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(ip.Metadata.Workspace),
		Name:      ip.Metadata.Name,
	}

	ip, err = r.client.NetworkV1.GetPublicIpUntilState(ctx, wref, r.retry.with(data.Retry).untilState(sdk.ResourceStateActive))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading public ip",
			"An error was encountered while waiting for the public ip to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags := publicIpToResourceModel(ctx, ip)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "public ip updated")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *PublicIpResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PublicIpResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "deleting public ip")

	ip := &sdk.PublicIp{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Tenant:    r.tenant,
			Workspace: data.WorkspaceId.ValueString(),
			Name:      data.Name.ValueString(),
		},
	}

	err := r.client.NetworkV1.DeletePublicIp(ctx, ip)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting public ip",
			"An error was encountered when deleting the public ip.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for public ip to be deleted")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(ip.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(ip.Metadata.Workspace),
		Name:      ip.Metadata.Name,
	}

	err = r.client.NetworkV1.WatchPublicIpUntilDeleted(ctx, wref, r.retry.with(data.Retry).observer())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading public ip",
			"An error was encountered while waiting for the public ip to become deleted.\nError: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "public ip deleted")
}

func publicIpFromModel(tenant string, data PublicIpResourceModel) *sdk.PublicIp {
	return &sdk.PublicIp{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Tenant:    tenant,
			Workspace: data.WorkspaceId.ValueString(),
			Name:      data.Name.ValueString(),
		},
		Labels:      toStringMap(data.Labels),
		Annotations: toStringMap(data.Annotations),
		Extensions:  toStringMap(data.Extensions),
		Spec: sdk.PublicIpSpec{
			Version: sdk.IPVersion(data.Version.ValueString()),
		},
	}
}

func publicIpToResourceModel(ctx context.Context, ip *sdk.PublicIp) (PublicIpResourceModel, diag.Diagnostics) {
	common, diags := publicIpToBaseModel(ctx, ip)
	return PublicIpResourceModel{publicIpModel: common}, diags
}
