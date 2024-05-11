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
var _ datasource.DataSource = &ProjectsDataSource{}

func NewProjectsDataSource() datasource.DataSource {
	return &ProjectsDataSource{}
}

// ProjectsDataSource defines the data source implementation.
type ProjectsDataSource struct {
	cli *cli.Cli
}

// ProjectsDataSourceModel describes the data source data model.
type ProjectsDataSourceModel struct {
	Projects []t.ProjectModel `tfsdk:"projects"`
}

func (d *ProjectsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_list"
}

func (d *ProjectsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "List of available Bitwarden projects",

		Attributes: map[string]schema.Attribute{
			"projects": schema.ListNestedAttribute{
				MarkdownDescription: "List of projects.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Unique identifier for the project.",
							Computed:            true,
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
				},
			},
		},
	}
}

func (d *ProjectsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectsDataSourceModel
	var jsonProjects []t.JSONProject

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	stdout, err := d.cli.ExecuteCommand("list", "projects")

	if err != nil {
		resp.Diagnostics.AddError("Bitwarden CLI encountered an error:", err.Error())
		return
	}

	err = json.Unmarshal(stdout, &jsonProjects)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to unmarshal JSON")
		return
	}

	data.Projects = make([]t.ProjectModel, len(jsonProjects))

	for projectIndex, jsonProject := range jsonProjects {
		data.Projects[projectIndex] = jsonProject.Parse()
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Fetched project values from Bitwarden Secrets CLI.")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
