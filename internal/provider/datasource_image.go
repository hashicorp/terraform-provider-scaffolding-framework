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
	_ datasource.DataSource              = (*ImageDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ImageDataSource)(nil)
)

type ImageDataSource struct {
	client *secapi.RegionalClient
	tenant string
}

func newImageDataSource() datasource.DataSource {
	return &ImageDataSource{}
}

func (d *ImageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

type ImageDataSourceModel struct {
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

	BlockStorageId  types.String `tfsdk:"block_storage_id"`
	CpuArchitecture types.String `tfsdk:"cpu_architecture"`
	Initializer     types.String `tfsdk:"initializer"`
	Boot            types.String `tfsdk:"boot"`

	State types.String `tfsdk:"state"`
}

func (d *ImageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Attributes: map[string]tfschema.Attribute{
			"id": tfschema.StringAttribute{
				Computed: true,
			},
			"name": tfschema.StringAttribute{
				Required: true,
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
			"block_storage_id": tfschema.StringAttribute{
				Computed: true,
			},
			"cpu_architecture": tfschema.StringAttribute{
				Computed: true,
			},
			"initializer": tfschema.StringAttribute{
				Computed: true,
			},
			"boot": tfschema.StringAttribute{
				Computed: true,
			},
			"state": tfschema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *ImageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured image data source")
}

func (d *ImageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ImageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "tenant_id", d.tenant)
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	tflog.Debug(ctx, "reading image data source")

	// Read the image

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(d.tenant),
		Name:   data.Name.ValueString(),
	}

	image, err := d.client.StorageV1.GetImage(ctx, tref)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading image",
			"An error was encountered when reading the image.\nError: "+err.Error(),
		)
		return
	}

	data, diags := imageToDataSourceModel(ctx, image)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func imageToDataSourceModel(ctx context.Context, image *sdk.Image) (ImageDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := ImageDataSourceModel{}
	model.Id = types.StringValue(image.Metadata.Ref)

	model.Name = types.StringValue(image.Metadata.Name)
	model.Tenant = types.StringValue(image.Metadata.Tenant)
	model.Region = types.StringValue(image.Metadata.Region)
	model.CreatedAt = fromTime(image.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(image.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(image.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, image.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, image.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, image.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	model.BlockStorageId = types.StringValue(image.Spec.BlockStorageRef.Resource)
	model.CpuArchitecture = types.StringValue(string(image.Spec.CpuArchitecture))
	model.Initializer = types.StringValue(string(image.Spec.Initializer))
	model.Boot = types.StringValue(string(image.Spec.Boot))

	model.State = types.StringValue(string(image.Status.State))

	return model, diags
}
