// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"terraform-provider-bitwarden-secrets/cli"
	t "terraform-provider-bitwarden-secrets/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &SecretDataSource{}

func NewSecretsDataSource() datasource.DataSource {
	return &SecretsDataSource{}
}

// SecretsDataSource defines the data source implementation.
type SecretsDataSource struct {
	cli *cli.Cli
}

// SecretsDataSourceModel describes the data source data model.
type SecretsDataSourceModel struct {
	Secrets []t.SecretModel `tfsdk:"secrets"`
}

func (d *SecretsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_list"
}

func (d *SecretsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "List of available Bitwarden secrets",

		Attributes: map[string]schema.Attribute{
			"secrets": schema.ListNestedAttribute{
				MarkdownDescription: "List of secrets.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Unique identifier for the secret.",
							Computed:            true,
						},
						"organization_id": schema.StringAttribute{
							MarkdownDescription: "Unique identifier for the organization associated with the secret.",
							Computed:            true,
						},
						"project_id": schema.StringAttribute{
							MarkdownDescription: "Unique identifier for the project associated with the secret.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: "Key identifying the secret.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "Value of the secret.",
							Computed:            true,
						},
						"note": schema.StringAttribute{
							MarkdownDescription: "Note included with the secret.",
							Computed:            true,
						},
						"creation_date": schema.StringAttribute{
							MarkdownDescription: "Date the secret was created.",
							Computed:            true,
						},
						"revision_date": schema.StringAttribute{
							MarkdownDescription: "Date the secret was last revised.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *SecretsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	cli, ok := req.ProviderData.(*cli.Cli)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected Cli, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.cli = cli
}

func (d *SecretsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecretsDataSourceModel
	var jsonSecrets []t.JSONSecret

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	stdout, err := d.cli.ExecuteCommand("list", "secrets")

	if err != nil {
		resp.Diagnostics.AddError("Bitwarden CLI encountered an error:", err.Error())
		return
	}

	err = json.Unmarshal(stdout, &jsonSecrets)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to unmarshal JSON")
		return
	}

	data.Secrets = make([]t.SecretModel, len(jsonSecrets))

	for secretIndex, jsonSecret := range jsonSecrets {
		data.Secrets[secretIndex] = jsonSecret.Parse()
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Fetched secret values from Bitwarden Secrets CLI.")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
