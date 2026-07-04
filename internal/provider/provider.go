package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type SecaProvider struct {
	Version string
}

type SecaProviderModel struct {
	Token  types.String `tfsdk:"token"`
	Tenant types.String `tfsdk:"tenant"`
	Region types.String `tfsdk:"region"`

	Retry *RetryModel `tfsdk:"retry"`

	GlobalProviders *SecaGlobalProvidersModel `tfsdk:"global_providers"`
}

type SecaGlobalProvidersModel struct {
	RegionV1        types.String `tfsdk:"region_v1"`
	AuthorizationV1 types.String `tfsdk:"authorization_v1"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SecaProvider{
			Version: version,
		}
	}
}

func (p *SecaProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "seca"
	resp.Version = p.Version
}

func (p *SecaProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Required: true,
			},
			"tenant": schema.StringAttribute{
				Required: true,
			},
			"region": schema.StringAttribute{
				Required: true,
			},
			"retry": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"delay": schema.NumberAttribute{
						Optional: true,
					},
					"interval": schema.NumberAttribute{
						Optional: true,
					},
					"max_attempts": schema.NumberAttribute{
						Optional: true,
					},
				},
			},
			"global_providers": schema.SingleNestedAttribute{
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

	tflog.Debug(ctx, "configuring seca provider")

	config := &clientConfig{
		Token:  model.Token.ValueString(),
		Tenant: model.Tenant.ValueString(),
		Region: model.Region.ValueString(),

		RetryDelay:       defaultRetryDelay,
		RetryInterval:    defaultRetryInterval,
		RetryMaxAttempts: defaultRetryMaxAttempts,
	}

	if model.Retry != nil {
		if !model.Retry.Delay.IsNull() {
			config.RetryDelay = numberToDuration(model.Retry.Delay)
		}
		if !model.Retry.Interval.IsNull() {
			config.RetryInterval = numberToDuration(model.Retry.Interval)
		}
		if !model.Retry.MaxAttempts.IsNull() {
			config.RetryMaxAttempts = numberToInt(model.Retry.MaxAttempts)
		}
	}

	config.GlobalProviders = &clientConfigGlobalProviders{
		RegionV1:        model.GlobalProviders.RegionV1.ValueString(),
		AuthorizationV1: model.GlobalProviders.AuthorizationV1.ValueString(),
	}

	clients, err := initClients(ctx, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to initialize SDK client",
			"Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = *clients
	resp.ResourceData = *clients

	tflog.Info(ctx, "configured seca provider")
}

func (p *SecaProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newWorkspaceResource,
		newImageResource,
		newBlockStorageResource,
		newNetworkResource,
		newInternetGatewayResource,
		newRouteTableResource,
		newSubnetResource,
		newSecurityGroupResource,
		newPublicIpResource,
		newNicResource,
	}
}

func (p *SecaProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newWorkspaceDataSource,
		newRegionDataSource,
		newImageDataSource,
		newBlockStorageDataSource,
		newStorageSkuDataSource,
		newNetworkSkuDataSource,
		newNetworkDataSource,
		newInternetGatewayDataSource,
		newRouteTableDataSource,
		newSubnetDataSource,
		newSecurityGroupDataSource,
		newPublicIpDataSource,
		newNicDataSource,
	}
}
