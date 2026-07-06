package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	tfschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
	"github.com/eu-sovereign-cloud/go-sdk/secapi"
)

var (
	_ resource.Resource                = (*SecurityGroupResource)(nil)
	_ resource.ResourceWithConfigure   = (*SecurityGroupResource)(nil)
	_ resource.ResourceWithImportState = (*SecurityGroupResource)(nil)
)

var sgPortsAttrTypes = map[string]attr.Type{
	"from": types.Int64Type,
	"to":   types.Int64Type,
	"list": types.ListType{ElemType: types.Int64Type},
}

var sgRuleAttrTypes = map[string]attr.Type{
	"direction":   types.StringType,
	"protocol":    types.StringType,
	"ports":       types.ObjectType{AttrTypes: sgPortsAttrTypes},
	"source_refs": types.ListType{ElemType: types.StringType},
}

type SecurityGroupResource struct {
	client *secapi.RegionalClient

	tenant string
	region string

	retry retryConfig
}

func newSecurityGroupResource() resource.Resource {
	return &SecurityGroupResource{}
}

func (r *SecurityGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

func (r *SecurityGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	workspaceID, name, ok := strings.Cut(req.ID, "/")
	if !ok || workspaceID == "" || name == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier in the format \"workspace_id/name\", got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), workspaceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
}

type SGPortsModel struct {
	From types.Int64 `tfsdk:"from"`
	To   types.Int64 `tfsdk:"to"`
	List types.List  `tfsdk:"list"`
}

type SGRuleModel struct {
	Direction  types.String `tfsdk:"direction"`
	Protocol   types.String `tfsdk:"protocol"`
	Ports      types.Object `tfsdk:"ports"`
	SourceRefs types.List   `tfsdk:"source_refs"`
}

type SecurityGroupResourceModel struct {
	securityGroupModel

	Retry *RetryModel `tfsdk:"retry"`
}

func (r *SecurityGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	portsAttrs := map[string]tfschema.Attribute{
		"from": tfschema.Int64Attribute{Optional: true},
		"to":   tfschema.Int64Attribute{Optional: true},
		"list": tfschema.ListAttribute{
			ElementType: types.Int64Type,
			Optional:    true,
		},
	}

	ruleAttrs := map[string]tfschema.Attribute{
		"direction": tfschema.StringAttribute{Required: true},
		"protocol":  tfschema.StringAttribute{Required: true},
		"ports": tfschema.SingleNestedAttribute{
			Optional:   true,
			Attributes: portsAttrs,
		},
		"source_refs": tfschema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
	}

	resp.Schema = tfschema.Schema{
		Attributes: map[string]tfschema.Attribute{
			"id": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": tfschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"workspace_id": tfschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tenant": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_provider": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deleted_at": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_modified_at": tfschema.StringAttribute{
				Computed: true,
			},
			"labels": tfschema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"annotations": tfschema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"extensions": tfschema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"rules": tfschema.ListNestedAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: tfschema.NestedAttributeObject{
					Attributes: ruleAttrs,
				},
			},
			"rule_refs": tfschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"retry": retryResourceSchema(),
		},
	}
}

func (r *SecurityGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = clients.RegionalClient
	r.tenant = clients.Tenant
	r.region = clients.Region
	r.retry = retryConfig{
		delay:       clients.RetryDelay,
		interval:    clients.RetryInterval,
		maxAttempts: clients.RetryMaxAttempts,
	}

	tflog.Debug(ctx, "configured security group resource")
}

func (r *SecurityGroupResource) logFields(ctx context.Context, data SecurityGroupResourceModel) context.Context {
	ctx = tflog.SetField(ctx, "tenant_id", r.tenant)
	ctx = tflog.SetField(ctx, "workspace_id", data.WorkspaceId.ValueString())
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	return ctx
}

func (r *SecurityGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecurityGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "creating security group")

	sg, diags := securityGroupFromModel(ctx, r.tenant, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sg, err := r.client.NetworkV1.CreateOrUpdateSecurityGroup(ctx, sg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating security group",
			"An error was encountered when creating the security group.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for security group to become active")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(sg.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(sg.Metadata.Workspace),
		Name:      sg.Metadata.Name,
	}

	sg, err = r.client.NetworkV1.GetSecurityGroupUntilState(ctx, wref, r.retry.with(data.Retry).untilState(sdk.ResourceStateActive))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading security group",
			"An error was encountered while waiting for the security group to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags := securityGroupToResourceModel(ctx, sg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "security group created")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *SecurityGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "reading security group")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(r.tenant),
		Workspace: secapi.WorkspaceID(data.WorkspaceId.ValueString()),
		Name:      data.Name.ValueString(),
	}

	sg, err := r.client.NetworkV1.GetSecurityGroup(ctx, wref)
	if err == secapi.ErrResourceNotFound {
		tflog.Debug(ctx, "security group not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error reading security group",
			"An error was encountered when reading the security group.\nError: "+err.Error(),
		)
		return
	}

	result, diags := securityGroupToResourceModel(ctx, sg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *SecurityGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SecurityGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "updating security group")

	sg, diags := securityGroupFromModel(ctx, r.tenant, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sg, err := r.client.NetworkV1.CreateOrUpdateSecurityGroup(ctx, sg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating security group",
			"An error was encountered when updating the security group.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for security group to become active")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(sg.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(sg.Metadata.Workspace),
		Name:      sg.Metadata.Name,
	}

	sg, err = r.client.NetworkV1.GetSecurityGroupUntilState(ctx, wref, r.retry.with(data.Retry).untilState(sdk.ResourceStateActive))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading security group",
			"An error was encountered while waiting for the security group to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags := securityGroupToResourceModel(ctx, sg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "security group updated")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *SecurityGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SecurityGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "deleting security group")

	sg := &sdk.SecurityGroup{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Tenant:    r.tenant,
			Workspace: data.WorkspaceId.ValueString(),
			Name:      data.Name.ValueString(),
		},
	}

	err := r.client.NetworkV1.DeleteSecurityGroup(ctx, sg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting security group",
			"An error was encountered when deleting the security group.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for security group to be deleted")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(sg.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(sg.Metadata.Workspace),
		Name:      sg.Metadata.Name,
	}

	err = r.client.NetworkV1.WatchSecurityGroupUntilDeleted(ctx, wref, r.retry.with(data.Retry).observer())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading security group",
			"An error was encountered while waiting for the security group to become deleted.\nError: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "security group deleted")
}

func securityGroupFromModel(ctx context.Context, tenant string, data SecurityGroupResourceModel) (*sdk.SecurityGroup, diag.Diagnostics) {
	var diags diag.Diagnostics

	rules, d := sgRulesFromModel(ctx, data.Rules)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	sg := &sdk.SecurityGroup{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Tenant:    tenant,
			Workspace: data.WorkspaceId.ValueString(),
			Name:      data.Name.ValueString(),
		},
		Labels:      toStringMap(data.Labels),
		Annotations: toStringMap(data.Annotations),
		Extensions:  toStringMap(data.Extensions),
		Spec: sdk.SecurityGroupSpec{
			Rules: rules,
		},
	}

	return sg, diags
}

func sgRulesFromModel(ctx context.Context, list types.List) ([]sdk.SecurityGroupRuleSpec, diag.Diagnostics) {
	var diags diag.Diagnostics

	if list.IsNull() || list.IsUnknown() {
		return nil, diags
	}

	var rules []SGRuleModel
	diags.Append(list.ElementsAs(ctx, &rules, false)...)
	if diags.HasError() {
		return nil, diags
	}

	specs := make([]sdk.SecurityGroupRuleSpec, 0, len(rules))
	for _, r := range rules {
		spec := sdk.SecurityGroupRuleSpec{
			Direction: sdk.SecurityGroupRuleSpecDirection(r.Direction.ValueString()),
			Protocol:  sdk.SecurityGroupRuleSpecProtocol(r.Protocol.ValueString()),
		}

		if !r.Ports.IsNull() && !r.Ports.IsUnknown() {
			attrs := r.Ports.Attributes()
			ports := &sdk.Ports{}
			if fromAttr, ok := attrs["from"].(types.Int64); ok && !fromAttr.IsNull() {
				ports.From = int(fromAttr.ValueInt64())
			}
			if toAttr, ok := attrs["to"].(types.Int64); ok && !toAttr.IsNull() {
				ports.To = int(toAttr.ValueInt64())
			}
			if listAttr, ok := attrs["list"].(types.List); ok && !listAttr.IsNull() {
				var portList []int64
				diags.Append(listAttr.ElementsAs(ctx, &portList, false)...)
				if diags.HasError() {
					return nil, diags
				}
				for _, p := range portList {
					ports.List = append(ports.List, int(p))
				}
			}
			spec.Ports = ports
		}

		if !r.SourceRefs.IsNull() && !r.SourceRefs.IsUnknown() {
			var refs []string
			diags.Append(r.SourceRefs.ElementsAs(ctx, &refs, false)...)
			if diags.HasError() {
				return nil, diags
			}
			for _, ref := range refs {
				spec.SourceRef = append(spec.SourceRef, sdk.Reference{Resource: ref})
			}
		}

		specs = append(specs, spec)
	}

	return specs, diags
}

func securityGroupToResourceModel(ctx context.Context, sg *sdk.SecurityGroup) (SecurityGroupResourceModel, diag.Diagnostics) {
	common, diags := securityGroupToBaseModel(ctx, sg)
	return SecurityGroupResourceModel{securityGroupModel: common}, diags
}

func sgRulesToListValue(ctx context.Context, specs []sdk.SecurityGroupRuleSpec) (types.List, diag.Diagnostics) {
	if len(specs) == 0 {
		return types.ListValueMust(types.ObjectType{AttrTypes: sgRuleAttrTypes}, []attr.Value{}), nil
	}

	elems := make([]attr.Value, 0, len(specs))
	for _, s := range specs {
		portsObj, d := sgPortsToObjectValue(ctx, s.Ports)
		if d.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: sgRuleAttrTypes}), d
		}

		sourceRefs := make([]attr.Value, 0, len(s.SourceRef))
		for _, ref := range s.SourceRef {
			sourceRefs = append(sourceRefs, types.StringValue(ref.Resource))
		}
		sourceRefsList, d := types.ListValue(types.StringType, sourceRefs)
		if d.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: sgRuleAttrTypes}), d
		}

		obj, d := types.ObjectValue(sgRuleAttrTypes, map[string]attr.Value{
			"direction":   types.StringValue(string(s.Direction)),
			"protocol":    types.StringValue(string(s.Protocol)),
			"ports":       portsObj,
			"source_refs": sourceRefsList,
		})
		if d.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: sgRuleAttrTypes}), d
		}
		elems = append(elems, obj)
	}

	return types.ListValue(types.ObjectType{AttrTypes: sgRuleAttrTypes}, elems)
}

func sgPortsToObjectValue(_ context.Context, ports *sdk.Ports) (types.Object, diag.Diagnostics) {
	if ports == nil {
		return types.ObjectNull(sgPortsAttrTypes), nil
	}

	portList := make([]attr.Value, 0, len(ports.List))
	for _, p := range ports.List {
		portList = append(portList, types.Int64Value(int64(p)))
	}

	listVal, diags := types.ListValue(types.Int64Type, portList)
	if diags.HasError() {
		return types.ObjectNull(sgPortsAttrTypes), diags
	}

	return types.ObjectValue(sgPortsAttrTypes, map[string]attr.Value{
		"from": types.Int64Value(int64(ports.From)),
		"to":   types.Int64Value(int64(ports.To)),
		"list": listVal,
	})
}

func sgRuleRefsToListValue(_ context.Context, refs []sdk.Reference) (types.List, diag.Diagnostics) {
	if len(refs) == 0 {
		return types.ListValueMust(types.StringType, []attr.Value{}), nil
	}

	elems := make([]attr.Value, 0, len(refs))
	for _, ref := range refs {
		elems = append(elems, types.StringValue(ref.Resource))
	}

	return types.ListValue(types.StringType, elems)
}
