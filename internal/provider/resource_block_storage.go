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
	_ resource.Resource                = (*BlockStorageResource)(nil)
	_ resource.ResourceWithConfigure   = (*BlockStorageResource)(nil)
	_ resource.ResourceWithImportState = (*BlockStorageResource)(nil)
)

type BlockStorageResource struct {
	client *secapi.RegionalClient

	tenant string
	region string

	retry retryConfig
}

func newBlockStorageResource() resource.Resource {
	return &BlockStorageResource{}
}

func (resource *BlockStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_storage"
}

func (r *BlockStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

type BlockStorageModel struct {
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

	SizeGB        types.Int64  `tfsdk:"size_gb"`
	SkuId         types.String `tfsdk:"sku_id"`
	SourceImageId types.String `tfsdk:"source_image_id"`

	Retry *RetryModel `tfsdk:"retry"`
}

func (resource *BlockStorageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"size_gb": tfschema.Int64Attribute{
				Required: true,
			},
			"sku_id": tfschema.StringAttribute{
				Required: true,
			},
			"source_image_id": tfschema.StringAttribute{
				Optional: true,
			},
			"retry": retryResourceSchema(),
		},
	}
}

func (r *BlockStorageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured block storage resource")
}

func (r *BlockStorageResource) logFields(ctx context.Context, data BlockStorageModel) context.Context {
	ctx = tflog.SetField(ctx, "tenant_id", r.tenant)
	ctx = tflog.SetField(ctx, "workspace_id", data.WorkspaceId.ValueString())
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	return ctx
}

func (resource *BlockStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BlockStorageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = resource.logFields(ctx, data)
	tflog.Debug(ctx, "creating block storage")

	// Create the block storage

	block := blockStorageFromModel(resource.tenant, data)

	block, err := resource.client.StorageV1.CreateOrUpdateBlockStorage(ctx, block)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating block storage",
			"An error was encountered when creating the block storage.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for block storage to become active")

	// Wait until it is active

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(block.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(block.Metadata.Workspace),
		Name:      block.Metadata.Name,
	}

	config := resource.retry.with(data.Retry).untilState(sdk.ResourceStateActive)

	block, err = resource.client.StorageV1.GetBlockStorageUntilState(ctx, wref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading block storage",
			"An error was encountered while waiting for the block storage to become active.\nError: "+err.Error(),
		)
		return
	}

	data, diags := blockStorageToResourceModel(ctx, block)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "block storage created")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (resource *BlockStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BlockStorageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = resource.logFields(ctx, data)
	tflog.Debug(ctx, "reading block storage")

	// Read the block storage

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(resource.tenant),
		Workspace: secapi.WorkspaceID(data.WorkspaceId.ValueString()),
		Name:      data.Name.ValueString(),
	}

	block, err := resource.client.StorageV1.GetBlockStorage(ctx, wref)
	if err == secapi.ErrResourceNotFound {
		tflog.Debug(ctx, "block storage not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error reading block storage",
			"An error was encountered when reading the block storage.\nError: "+err.Error(),
		)
		return
	}

	data, diags := blockStorageToResourceModel(ctx, block)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (resource *BlockStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BlockStorageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = resource.logFields(ctx, data)
	tflog.Debug(ctx, "updating block storage")

	// Update the block storage

	block := blockStorageFromModel(resource.tenant, data)

	block, err := resource.client.StorageV1.CreateOrUpdateBlockStorage(ctx, block)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating block storage",
			"An error was encountered when updating the block storage.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for block storage to become active")

	// Wait until it is active

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(block.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(block.Metadata.Workspace),
		Name:      block.Metadata.Name,
	}

	config := resource.retry.with(data.Retry).untilState(sdk.ResourceStateActive)

	block, err = resource.client.StorageV1.GetBlockStorageUntilState(ctx, wref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading block storage",
			"An error was encountered while waiting for the block storage to become active.\nError: "+err.Error(),
		)
		return
	}

	data, diags := blockStorageToResourceModel(ctx, block)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "block storage updated")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (resource *BlockStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BlockStorageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = resource.logFields(ctx, data)
	tflog.Debug(ctx, "deleting block storage")

	// Delete the block storage

	block := &sdk.BlockStorage{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Tenant:    resource.tenant,
			Workspace: data.WorkspaceId.ValueString(),
			Name:      data.Name.ValueString(),
		},
	}

	err := resource.client.StorageV1.DeleteBlockStorage(ctx, block)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting block storage",
			"An error was encountered when deleting the block storage.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for block storage to be deleted")

	// Wait until it is deleted

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(block.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(block.Metadata.Workspace),
		Name:      block.Metadata.Name,
	}

	config := resource.retry.with(data.Retry).observer()

	err = resource.client.StorageV1.WatchBlockStorageUntilDeleted(ctx, wref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading block storage",
			"An error was encountered while waiting for the block storage to become deleted.\nError: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "block storage deleted")
}

func blockStorageFromModel(tenant string, data BlockStorageModel) *sdk.BlockStorage {
	block := &sdk.BlockStorage{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Tenant:    tenant,
			Workspace: data.WorkspaceId.ValueString(),
			Name:      data.Name.ValueString(),
		},
		Labels:      toStringMap(data.Labels),
		Annotations: toStringMap(data.Annotations),
		Extensions:  toStringMap(data.Extensions),
		Spec: sdk.BlockStorageSpec{
			SizeGB: int(data.SizeGB.ValueInt64()),
			SkuRef: sdk.Reference{
				Resource: data.SkuId.ValueString(),
			},
		},
	}

	if !data.SourceImageId.IsNull() && !data.SourceImageId.IsUnknown() {
		block.Spec.SourceImageRef = &sdk.Reference{
			Resource: data.SourceImageId.ValueString(),
		}
	}

	return block
}

func blockStorageToResourceModel(ctx context.Context, block *sdk.BlockStorage) (BlockStorageModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := BlockStorageModel{}
	model.Id = types.StringValue(block.Metadata.Ref)

	model.Name = types.StringValue(block.Metadata.Name)
	model.WorkspaceId = types.StringValue(block.Metadata.Workspace)
	model.Tenant = types.StringValue(block.Metadata.Tenant)
	model.Region = types.StringValue(block.Metadata.Region)
	model.ResourceProvider = refToResourceProvider(block.Metadata.Ref)
	model.CreatedAt = fromTime(block.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(block.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(block.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, block.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, block.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, block.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	model.SizeGB = types.Int64Value(int64(block.Spec.SizeGB))
	model.SkuId = types.StringValue(block.Spec.SkuRef.Resource)
	model.SourceImageId = fromRefPtr(block.Spec.SourceImageRef)

	return model, diags
}
