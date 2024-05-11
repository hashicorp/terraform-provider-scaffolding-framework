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
var _ datasource.DataSource = &ProjectDataSource{}

func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

// ProjectDataSource defines the data source implementation.
type ProjectDataSource struct {
	cli *cli.Cli
}

// ProjectDataSourceModel describes the data source data model.

func (d *ProjectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *ProjectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Bitwarden Project",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the project.",
				Required:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the organization.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the project.",
				Computed:            true,
			},
			"creation_date": schema.StringAttribute{
				MarkdownDescription: "Date the project was created.",
				Computed:            true,
			},
			"revision_date": schema.StringAttribute{
				MarkdownDescription: "Date the project was last revised.",
				Computed:            true,
			},
		},
	}
}

func (d *ProjectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	cli, ok := req.ProviderData.(*cli.Cli)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected string, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.cli = cli
}

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data t.ProjectModel
	var jsonProject t.JSONProject

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	stdout, err := d.cli.ExecuteCommand("project", "get", data.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Bitwarden CLI encountered an error:", err.Error())
		return
	}

	err = json.Unmarshal(stdout, &jsonProject)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to unmarshal JSON")
		return
	}

	data = jsonProject.Parse()

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Fetched project value from Bitwarden Secrets CLI.")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
