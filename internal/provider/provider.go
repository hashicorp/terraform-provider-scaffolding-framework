package provider

import (
	"context"

	"github.com/eu-sovereign-cloud/terraform-provider-seca/internal/sdk"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SecaProvider struct {
	version string
}

type SecaProviderModel struct {
	Token     types.String        `tfsdk:"token"`
	Region    types.String        `tfsdk:"region"`
	Providers *SecaProvidersModel `tfsdk:"providers"`
}

type SecaProvidersModel struct {
	RegionV1        types.String `tfsdk:"region_v1"`
	AuthorizationV1 types.String `tfsdk:"authorization_v1"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SecaProvider{
			version: version,
		}
	}
}

func (p *SecaProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "seca"
	resp.Version = p.version
}

func (p *SecaProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Required: true,
			},
			"region": schema.StringAttribute{
				Required: true,
			},
			"providers": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"region_v1": schema.StringAttribute{
						Required: true,
					},
					"authorization_v1": schema.StringAttribute{
						Optional: true,
					},
				},
			},
		},
	}
}

func (p *SecaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var model SecaProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := &sdk.Config{
		Token:  model.Token.ValueString(),
		Region: model.Region.ValueString(),
	}

	if model.Providers != nil {
		config.GlobalProviders = &sdk.ConfigGlobalProviders{
			RegionV1:        model.Providers.RegionV1.ValueString(),
			AuthorizationV1: model.Providers.AuthorizationV1.ValueString(),
		}
	}

	clients, err := sdk.InitClients(ctx, config)
	if err != nil {
		resp.Diagnostics.AddError("Unable to initialize SDK client",
			"Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = *clients
	resp.ResourceData = *clients
}

func (p *SecaProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *SecaProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
