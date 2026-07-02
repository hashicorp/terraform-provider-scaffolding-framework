package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	tfschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
	"github.com/eu-sovereign-cloud/go-sdk/secapi"
)

var (
	_ datasource.DataSource              = (*RegionDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*RegionDataSource)(nil)
)

type RegionDataSource struct {
	client *secapi.GlobalClient
}

func newRegionDataSource() datasource.DataSource {
	return &RegionDataSource{}
}

func (d *RegionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_region"
}

type RegionProviderModel struct {
	Name    types.String `tfsdk:"name"`
	Version types.String `tfsdk:"version"`
	Url     types.String `tfsdk:"url"`
}

var regionProviderAttrTypes = map[string]attr.Type{
	"name":    types.StringType,
	"version": types.StringType,
	"url":     types.StringType,
}

type RegionDataSourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	CreatedAt      types.String `tfsdk:"created_at"`
	DeletedAt      types.String `tfsdk:"deleted_at"`
	LastModifiedAt types.String `tfsdk:"last_modified_at"`

	AvailableZones types.List `tfsdk:"available_zones"`
	Providers      types.List `tfsdk:"providers"`
}

func (d *RegionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Attributes: map[string]tfschema.Attribute{
			"id": tfschema.StringAttribute{
				Computed: true,
			},
			"name": tfschema.StringAttribute{
				Required: true,
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
			"available_zones": tfschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"providers": tfschema.ListNestedAttribute{
				Computed: true,
				NestedObject: tfschema.NestedAttributeObject{
					Attributes: map[string]tfschema.Attribute{
						"name": tfschema.StringAttribute{
							Computed: true,
						},
						"version": tfschema.StringAttribute{
							Computed: true,
						},
						"url": tfschema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *RegionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = clients.GlobalClient

	tflog.Debug(ctx, "configured region data source")
}

func (d *RegionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RegionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	tflog.Debug(ctx, "reading region data source")

	// Read the region

	region, err := d.client.RegionV1.GetRegion(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading region",
			"An error was encountered when reading the region.\nError: "+err.Error(),
		)
		return
	}

	data, diags := regionToDataSourceModel(ctx, region)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func regionToDataSourceModel(ctx context.Context, region *sdk.Region) (RegionDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := RegionDataSourceModel{}
	model.Id = types.StringValue(region.Metadata.Ref)

	model.Name = types.StringValue(region.Metadata.Name)
	model.CreatedAt = fromTime(region.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(region.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(region.Metadata.LastModifiedAt)

	availableZones, d := types.ListValueFrom(ctx, types.StringType, region.Spec.AvailableZones)
	diags.Append(d...)
	model.AvailableZones = availableZones

	providers := make([]RegionProviderModel, 0, len(region.Spec.Providers))
	for _, p := range region.Spec.Providers {
		providers = append(providers, RegionProviderModel{
			Name:    types.StringValue(p.Name),
			Version: types.StringValue(p.Version),
			Url:     types.StringValue(p.Url),
		})
	}

	providersList, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: regionProviderAttrTypes}, providers)
	diags.Append(d...)
	model.Providers = providersList

	return model, diags
}
