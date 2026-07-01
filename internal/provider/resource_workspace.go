package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	tfschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
	"github.com/eu-sovereign-cloud/go-sdk/secapi"
)

var (
	_ resource.Resource              = (*WorkspaceResource)(nil)
	_ resource.ResourceWithConfigure = (*WorkspaceResource)(nil)
)

type WorkspaceResource struct {
	client *secapi.RegionalClient

	tenant string
	region string

	retryDelay       time.Duration
	retryInterval    time.Duration
	retryMaxAttempts int
}

func newWorkspaceResource() resource.Resource {
	return &WorkspaceResource{}
}

func (resource *WorkspaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

type WorkspaceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Tenant         types.String `tfsdk:"tenant"`
	Region         types.String `tfsdk:"region"`
	CreatedAt      types.String `tfsdk:"created_at"`
	DeletedAt      types.String `tfsdk:"deleted_at"`
	LastModifiedAt types.String `tfsdk:"last_modified_at"`

	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
	Extensions  types.Map `tfsdk:"extensions"`
}

func (resource *WorkspaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Attributes: map[string]tfschema.Attribute{
			"id": tfschema.StringAttribute{
				Computed: true,
			},
			"name": tfschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tenant": tfschema.StringAttribute{
				Computed: true,
			},
			"region": tfschema.StringAttribute{
				Computed: true,
			},
			"created_at": tfschema.StringAttribute{
				Computed: true,
			},
			"deleted_at": tfschema.StringAttribute{
				Computed: true,
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
		},
	}
}

func (r *WorkspaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.retryDelay = clients.RetryDelay
	r.retryInterval = clients.RetryInterval
	r.retryMaxAttempts = clients.RetryMaxAttempts
}

func (resource *WorkspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WorkspaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the workspace

	workspace := &sdk.Workspace{
		Metadata: &sdk.RegionalResourceMetadata{
			Tenant: resource.tenant,
			Name:   data.Name.ValueString(),
		},
		Labels:      toStringMap(data.Labels),
		Annotations: toStringMap(data.Annotations),
		Extensions:  toStringMap(data.Extensions),
	}

	workspace, err := resource.client.WorkspaceV1.CreateOrUpdateWorkspace(ctx, workspace)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating workspace",
			"An error was encountered when creating the workspace.\nError: "+err.Error(),
		)
		return
	}

	// Wait until it is active

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(workspace.Metadata.Tenant),
		Name:   workspace.Metadata.Name,
	}

	config := secapi.ResourceObserverUntilValueConfig[sdk.ResourceState]{
		ExpectedValues: []sdk.ResourceState{sdk.ResourceStateActive},
		Delay:          resource.retryDelay,
		Interval:       resource.retryInterval,
		MaxAttempts:    resource.retryMaxAttempts,
	}

	workspace, err = resource.client.WorkspaceV1.GetWorkspaceUntilState(ctx, tref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading workspace",
			"An error was encountered while waiting for the workspace to become active.\nError: "+err.Error(),
		)
		return
	}

	data, diags := workspaceToResourceModel(ctx, workspace)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (resource *WorkspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WorkspaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the workspace

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(resource.tenant),
		Name:   data.Name.ValueString(),
	}

	workspace, err := resource.client.WorkspaceV1.GetWorkspace(ctx, tref)
	if err == secapi.ErrResourceNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error reading workspace",
			"An error was encountered when reading the workspace.\nError: "+err.Error(),
		)
		return
	}

	data, diags := workspaceToResourceModel(ctx, workspace)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (resource *WorkspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WorkspaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the workspace

	workspace := &sdk.Workspace{
		Metadata: &sdk.RegionalResourceMetadata{
			Tenant: resource.tenant,
			Name:   data.Name.ValueString(),
		},
		Labels:      toStringMap(data.Labels),
		Annotations: toStringMap(data.Annotations),
		Extensions:  toStringMap(data.Extensions),
	}

	workspace, err := resource.client.WorkspaceV1.CreateOrUpdateWorkspace(ctx, workspace)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating workspace",
			"An error was encountered when updating the workspace.\nError: "+err.Error(),
		)
		return
	}

	// Wait until it is active

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(workspace.Metadata.Tenant),
		Name:   workspace.Metadata.Name,
	}

	config := secapi.ResourceObserverUntilValueConfig[sdk.ResourceState]{
		ExpectedValues: []sdk.ResourceState{sdk.ResourceStateActive},
		Delay:          resource.retryDelay,
		Interval:       resource.retryInterval,
		MaxAttempts:    resource.retryMaxAttempts,
	}

	workspace, err = resource.client.WorkspaceV1.GetWorkspaceUntilState(ctx, tref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading workspace",
			"An error was encountered while waiting for the workspace to become active.\nError: "+err.Error(),
		)
		return
	}

	data, diags := workspaceToResourceModel(ctx, workspace)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (resource *WorkspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WorkspaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the workspace

	workspace := &sdk.Workspace{
		Metadata: &sdk.RegionalResourceMetadata{
			Tenant: resource.tenant,
			Name:   data.Name.ValueString(),
		},
	}

	err := resource.client.WorkspaceV1.DeleteWorkspace(ctx, workspace)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting workspace",
			"An error was encountered when deleting the workspace.\nError: "+err.Error(),
		)
		return
	}

	// Wait until it is deleted

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(workspace.Metadata.Tenant),
		Name:   workspace.Metadata.Name,
	}

	config := secapi.ResourceObserverConfig{
		Delay:       resource.retryDelay,
		Interval:    resource.retryInterval,
		MaxAttempts: resource.retryMaxAttempts,
	}

	err = resource.client.WorkspaceV1.WatchWorkspaceUntilDeleted(ctx, tref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading workspace",
			"An error was encountered while waiting for the workspace to become deleted.\nError: "+err.Error(),
		)
		return
	}
}

func workspaceToResourceModel(ctx context.Context, workspace *sdk.Workspace) (WorkspaceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := WorkspaceModel{}
	model.Id = types.StringValue(workspace.Metadata.Ref)

	model.Name = types.StringValue(workspace.Metadata.Name)
	model.Tenant = types.StringValue(workspace.Metadata.Tenant)
	model.Region = types.StringValue(workspace.Metadata.Region)
	model.CreatedAt = fromTime(workspace.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(workspace.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(workspace.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, workspace.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, workspace.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, workspace.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	return model, diags
}
