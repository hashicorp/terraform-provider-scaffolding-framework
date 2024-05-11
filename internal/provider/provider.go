// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"terraform-provider-bitwarden-secrets/cli"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &BitwardenSecretsProvider{}
var _ provider.ProviderWithFunctions = &BitwardenSecretsProvider{}

// BitwardenSecretsProvider defines the provider implementation.
type BitwardenSecretsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ScaffoldingProviderModel describes the provider data model.
type ScaffoldingProviderModel struct {
	AccessToken types.String `tfsdk:"access_token"`
	ServerURL   types.String `tfsdk:"server_url"`
}

func (p *BitwardenSecretsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bitwarden-secrets"
	resp.Version = p.version
}

func (p *BitwardenSecretsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_token": schema.StringAttribute{
				MarkdownDescription: "Token used to authenticate with the Bitwarden Secrets CLI.",
				Required:            true,
				Sensitive:           true,
			},
			"server_url": schema.StringAttribute{
				MarkdownDescription: "URL of the Bitwarden server.",
				Optional:            true,
			},
		},
	}
}

func (p *BitwardenSecretsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ScaffoldingProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	cli := cli.NewCli(data.AccessToken.ValueString(), data.ServerURL.ValueString())

	resp.DataSourceData = cli
	resp.ResourceData = cli
}

func (p *BitwardenSecretsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSecretResource,
		NewProjectResource,
	}
}

func (p *BitwardenSecretsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSecretDataSource,
		NewProjectDataSource,
		NewSecretsDataSource,
		NewProjectsDataSource,
	}
}

func (p *BitwardenSecretsProvider) Functions(ctx context.Context) []func() function.Function {
	return nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &BitwardenSecretsProvider{
			version: version,
		}
	}
}
