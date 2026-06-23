package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	tfschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
	"github.com/eu-sovereign-cloud/go-sdk/secapi"
)

var (
	_ datasource.DataSource              = (*StorageSkuDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*StorageSkuDataSource)(nil)
)

type StorageSkuDataSource struct {
	client *secapi.RegionalClient
	tenant string
}

func newStorageSkuDataSource() datasource.DataSource {
	return &StorageSkuDataSource{}
}

func (d *StorageSkuDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_sku"
}

type StorageSkuDataSourceModel struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Tenant types.String `tfsdk:"tenant"`
	Region types.String `tfsdk:"region"`

	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
	Extensions  types.Map `tfsdk:"extensions"`

	Iops          types.Int64  `tfsdk:"iops"`
	Type          types.String `tfsdk:"type"`
	MinVolumeSize types.Int64  `tfsdk:"min_volume_size"`
}

func (d *StorageSkuDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"iops": tfschema.Int64Attribute{
				Computed: true,
			},
			"type": tfschema.StringAttribute{
				Computed: true,
			},
			"min_volume_size": tfschema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *StorageSkuDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
}

func (d *StorageSkuDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StorageSkuDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the storage SKU

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(d.tenant),
		Name:   data.Name.ValueString(),
	}

	sku, err := d.client.StorageV1.GetSku(ctx, tref)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading storage SKU",
			"An error was encountered when reading the storage SKU.\nError: "+err.Error(),
		)
		return
	}

	data, diags := storageSkuToDataSourceModel(ctx, sku)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func storageSkuToDataSourceModel(ctx context.Context, sku *sdk.StorageSku) (StorageSkuDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := StorageSkuDataSourceModel{}
	model.Id = types.StringValue(sku.Metadata.Ref)

	model.Name = types.StringValue(sku.Metadata.Name)
	model.Tenant = types.StringValue(sku.Metadata.Tenant)
	model.Region = types.StringValue(sku.Metadata.Region)

	labels, d := fromStringMap(ctx, sku.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, sku.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, sku.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	if sku.Spec != nil {
		model.Iops = types.Int64Value(int64(sku.Spec.Iops))
		model.Type = types.StringValue(string(sku.Spec.Type))
		model.MinVolumeSize = types.Int64Value(int64(sku.Spec.MinVolumeSize))
	}

	return model, diags
}
