// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"terraform-provider-bitwarden-secrets/cli"
	t "terraform-provider-bitwarden-secrets/types"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SecretResource{}
var _ resource.ResourceWithImportState = &SecretResource{}

func NewSecretResource() resource.Resource {
	return &SecretResource{}
}

// SecretResource defines the resource implementation.
type SecretResource struct {
	cli *cli.Cli
}

func (r *SecretResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *SecretResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Bitwarden Secret",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the secret.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the organization associated with the secret.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the project associated with the secret.",
				Required:            true,
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "Key identifying the secret.",
				Required:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Value of the secret.",
				Required:            true,
				Sensitive:           true,
			},
			"note": schema.StringAttribute{
				MarkdownDescription: "Note included with the secret.",
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString(""),
			},
			"creation_date": schema.StringAttribute{
				MarkdownDescription: "Date the secret was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"revision_date": schema.StringAttribute{
				MarkdownDescription: "Date the secret was last revised.",
				Computed:            true,
			},
		},
	}
}

func (r *SecretResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	cli, ok := req.ProviderData.(*cli.Cli)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Cli, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.cli = cli
}

func (r *SecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data t.SecretModel
	var jsonSecret t.JSONSecret

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ProjectID.ValueString() == "" {
		resp.Diagnostics.AddError("Project ID is required", "Project ID must be provided")
		return
	}
	if data.Key.ValueString() == "" {
		resp.Diagnostics.AddError("Key is required", "Key must be provided")
		return
	}

	stdout, err := r.cli.ExecuteCommand("secret", "create", "--note", data.Note.ValueString(), data.Key.ValueString(), data.Value.ValueString(), data.ProjectID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Bitwarden CLI encountered an error:", err.Error())
		return
	}

	err = json.Unmarshal(stdout, &jsonSecret)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to unmarshal JSON")
		return
	}

	data = jsonSecret.Parse()

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data t.SecretModel
	var jsonSecret t.JSONSecret

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	stdout, err := r.cli.ExecuteCommand("secret", "get", data.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Bitwarden CLI encountered an error:", err.Error())
		return
	}

	err = json.Unmarshal(stdout, &jsonSecret)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to unmarshal JSON")
		return
	}

	data = jsonSecret.Parse()

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data t.SecretModel
	var jsonSecret t.JSONSecret

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	arguments := []string{"secret", "edit"}

	key := data.Key.ValueString()
	if key != "" {
		arguments = append(arguments, "--key", key)
	}

	value := data.Value.ValueString()
	arguments = append(arguments, "--value", value)

	note := data.Note.ValueString()
	arguments = append(arguments, "--note", note)

	project_id := data.ProjectID.ValueString()
	if project_id != "" {
		arguments = append(arguments, "--project-id", project_id)
	}

	arguments = append(arguments, data.ID.ValueString())

	stdout, err := r.cli.ExecuteCommand(arguments...)

	if err != nil {
		resp.Diagnostics.AddError("Bitwarden CLI encountered an error:", err.Error())
		return
	}

	err = json.Unmarshal(stdout, &jsonSecret)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to unmarshal JSON")
		return
	}

	data = jsonSecret.Parse()

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data t.SecretModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.cli.ExecuteCommand("secret", "delete", data.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Bitwarden CLI encountered an error:", err.Error())
		return
	}

}

func (r *SecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
