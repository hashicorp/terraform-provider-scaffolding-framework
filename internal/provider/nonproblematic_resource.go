package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &NonproblematicResource{}

func NewNonproblematicResource() resource.Resource {
	return &NonproblematicResource{}
}

type NonproblematicResource struct {
	client *http.Client
}

type ConfigModel struct {
	Attr1 types.Int64 `tfsdk:""attr1`
	Attr2 types.Int64 `tfsdk:""attr2`
}

type TheAttributeModel struct {
	Type	types.String `tfsdk:"type"`
	Config *ConfigModel `tfsdk:"config"`
}

type NonproblematicResourceModel struct {
	Id                  types.String `tfsdk:"id"`
	Target              types.String `tfsdk:"target"`
	SecondaryName       types.String `tfsdk:"secondary_name"`
	TheAttribute	*TheAttributeModel `tfsdk:"the_attribute"`
	CreationTimestamp   types.String `tfsdk:"creation_timestamp"`
	LastUpdateTimestamp types.String `tfsdk:"last_update_timestamp"`
}

func (r *NonproblematicResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nonproblematic"
}

func (r *NonproblematicResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Nonproblematic resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Example identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"target": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"secondary_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"the_attribute": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"config": schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"attr1": schema.Int64Attribute{
								Optional: true,
								Computed: true,
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.UseStateForUnknown(),
								}
							},
							"attr2": schema.Int64Attribute{
								Optional: true,
								Computed: true,
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.UseStateForUnknown(),
								}
							},
						},
					}
				},
			},
			"creation_timestamp": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_update_timestamp": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *NonproblematicResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *NonproblematicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NonproblematicResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// some `POST` request done here

	data.Id = types.StringValue("example-id")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NonproblematicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NonproblematicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue("example-id")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NonproblematicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NonproblematicResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue("example-id")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NonproblematicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NonproblematicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
