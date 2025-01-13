// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ ephemeral.EphemeralResource = &ExampleEphemeralResource{}

func NewExampleEphemeralResource() ephemeral.EphemeralResource {
	return &ExampleEphemeralResource{}
}

// ExampleEphemeralResource defines the ephemeral resource implementation.
type ExampleEphemeralResource struct {
	// client *http.Client // If applicable, a client can be initialized here.
}

// ExampleEphemeralResourceModel describes the ephemeral resource data model.
type ExampleEphemeralResourceModel struct {
	ConfigurableAttribute types.String `tfsdk:"configurable_attribute"`
	Value                 types.String `tfsdk:"value"`
}

func (r *ExampleEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_example"
}

func (r *ExampleEphemeralResource) Schema(ctx context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example ephemeral resource",

		Attributes: map[string]schema.Attribute{
			"configurable_attribute": schema.StringAttribute{
				MarkdownDescription: "Example configurable attribute",
				Required:            true, // Ephemeral resources expect their dependencies to already exist.
			},
			"value": schema.StringAttribute{
				Computed: true,
				// Sensitive:           true, // If applicable, mark the attribute as sensitive.
				MarkdownDescription: "Example value",
			},
		},
	}
}

func (r *ExampleEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data ExampleEphemeralResourceModel

	// Read Terraform config data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }
	//
	// However, this example hardcodes setting the token attribute to a specific value for brevity.
	data.Value = types.StringValue("token-123")

	// Save data into ephemeral result data
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
