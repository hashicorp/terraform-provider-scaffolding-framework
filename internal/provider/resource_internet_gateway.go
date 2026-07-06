package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	tfschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
	"github.com/eu-sovereign-cloud/go-sdk/secapi"
)

var (
	_ resource.Resource                = (*InternetGatewayResource)(nil)
	_ resource.ResourceWithConfigure   = (*InternetGatewayResource)(nil)
	_ resource.ResourceWithImportState = (*InternetGatewayResource)(nil)
)

type InternetGatewayResource struct {
	client *secapi.RegionalClient

	tenant string
	region string

	retry retryConfig
}

func newInternetGatewayResource() resource.Resource {
	return &InternetGatewayResource{}
}

func (r *InternetGatewayResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_internet_gateway"
}

func (r *InternetGatewayResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

type InternetGatewayResourceModel struct {
	internetGatewayModel

	Retry *RetryModel `tfsdk:"retry"`
}

func (r *InternetGatewayResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"egress_only": tfschema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"retry": retryResourceSchema(),
		},
	}
}

func (r *InternetGatewayResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured internet gateway resource")
}

func (r *InternetGatewayResource) logFields(ctx context.Context, data InternetGatewayResourceModel) context.Context {
	ctx = tflog.SetField(ctx, "tenant_id", r.tenant)
	ctx = tflog.SetField(ctx, "workspace_id", data.WorkspaceId.ValueString())
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	return ctx
}

func (r *InternetGatewayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InternetGatewayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "creating internet gateway")

	gtw := internetGatewayFromModel(r.tenant, data)

	gtw, err := r.client.NetworkV1.CreateOrUpdateInternetGateway(ctx, gtw)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating internet gateway",
			"An error was encountered when creating the internet gateway.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for internet gateway to become active")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(gtw.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(gtw.Metadata.Workspace),
		Name:      gtw.Metadata.Name,
	}

	gtw, err = r.client.NetworkV1.GetInternetGatewayUntilState(ctx, wref, r.retry.with(data.Retry).untilState(sdk.ResourceStateActive))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading internet gateway",
			"An error was encountered while waiting for the internet gateway to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags := internetGatewayToResourceModel(ctx, gtw)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "internet gateway created")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *InternetGatewayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InternetGatewayResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "reading internet gateway")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(r.tenant),
		Workspace: secapi.WorkspaceID(data.WorkspaceId.ValueString()),
		Name:      data.Name.ValueString(),
	}

	gtw, err := r.client.NetworkV1.GetInternetGateway(ctx, wref)
	if err == secapi.ErrResourceNotFound {
		tflog.Debug(ctx, "internet gateway not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error reading internet gateway",
			"An error was encountered when reading the internet gateway.\nError: "+err.Error(),
		)
		return
	}

	result, diags := internetGatewayToResourceModel(ctx, gtw)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *InternetGatewayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data InternetGatewayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "updating internet gateway")

	gtw := internetGatewayFromModel(r.tenant, data)

	gtw, err := r.client.NetworkV1.CreateOrUpdateInternetGateway(ctx, gtw)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating internet gateway",
			"An error was encountered when updating the internet gateway.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for internet gateway to become active")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(gtw.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(gtw.Metadata.Workspace),
		Name:      gtw.Metadata.Name,
	}

	gtw, err = r.client.NetworkV1.GetInternetGatewayUntilState(ctx, wref, r.retry.with(data.Retry).untilState(sdk.ResourceStateActive))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading internet gateway",
			"An error was encountered while waiting for the internet gateway to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags := internetGatewayToResourceModel(ctx, gtw)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "internet gateway updated")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *InternetGatewayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InternetGatewayResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "deleting internet gateway")

	gtw := &sdk.InternetGateway{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Tenant:    r.tenant,
			Workspace: data.WorkspaceId.ValueString(),
			Name:      data.Name.ValueString(),
		},
	}

	err := r.client.NetworkV1.DeleteInternetGateway(ctx, gtw)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting internet gateway",
			"An error was encountered when deleting the internet gateway.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for internet gateway to be deleted")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(gtw.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(gtw.Metadata.Workspace),
		Name:      gtw.Metadata.Name,
	}

	err = r.client.NetworkV1.WatchInternetGatewayUntilDeleted(ctx, wref, r.retry.with(data.Retry).observer())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading internet gateway",
			"An error was encountered while waiting for the internet gateway to become deleted.\nError: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "internet gateway deleted")
}

func internetGatewayFromModel(tenant string, data InternetGatewayResourceModel) *sdk.InternetGateway {
	gtw := &sdk.InternetGateway{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Tenant:    tenant,
			Workspace: data.WorkspaceId.ValueString(),
			Name:      data.Name.ValueString(),
		},
		Labels:      toStringMap(data.Labels),
		Annotations: toStringMap(data.Annotations),
		Extensions:  toStringMap(data.Extensions),
		Spec: sdk.InternetGatewaySpec{
			EgressOnly: data.EgressOnly.ValueBool(),
		},
	}
	return gtw
}

func internetGatewayToResourceModel(ctx context.Context, gtw *sdk.InternetGateway) (InternetGatewayResourceModel, diag.Diagnostics) {
	common, diags := internetGatewayToBaseModel(ctx, gtw)
	return InternetGatewayResourceModel{internetGatewayModel: common}, diags
}
