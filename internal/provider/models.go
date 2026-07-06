package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

type networkModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	WorkspaceId      types.String `tfsdk:"workspace_id"`
	Tenant           types.String `tfsdk:"tenant"`
	Region           types.String `tfsdk:"region"`
	ResourceProvider types.String `tfsdk:"resource_provider"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DeletedAt        types.String `tfsdk:"deleted_at"`
	LastModifiedAt   types.String `tfsdk:"last_modified_at"`

	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
	Extensions  types.Map `tfsdk:"extensions"`

	SkuId           types.String     `tfsdk:"sku_id"`
	Cidr            NetworkCidrModel `tfsdk:"cidr"`
	AdditionalCidrs types.List       `tfsdk:"additional_cidrs"`
}

func networkToBaseModel(ctx context.Context, net *sdk.Network) (networkModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := networkModel{}
	model.Id = types.StringValue(net.Metadata.Ref)
	model.Name = types.StringValue(net.Metadata.Name)
	model.WorkspaceId = types.StringValue(net.Metadata.Workspace)
	model.Tenant = types.StringValue(net.Metadata.Tenant)
	model.Region = types.StringValue(net.Metadata.Region)
	model.ResourceProvider = refToResourceProvider(net.Metadata.Ref)
	model.CreatedAt = fromTime(net.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(net.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(net.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, net.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, net.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, net.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	model.SkuId = types.StringValue(net.Spec.SkuRef.Resource)

	if net.Status != nil {
		model.Cidr = cidrFromSDK(net.Status.Cidr)
		additionalCidrs, d := fromCidrList(ctx, net.Status.AdditionalCidrs)
		diags.Append(d...)
		model.AdditionalCidrs = additionalCidrs
	} else {
		model.Cidr = cidrFromSDK(net.Spec.Cidr)
		additionalCidrs, d := fromCidrList(ctx, net.Spec.AdditionalCidrs)
		diags.Append(d...)
		model.AdditionalCidrs = additionalCidrs
	}

	return model, diags
}

type securityGroupModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	WorkspaceId      types.String `tfsdk:"workspace_id"`
	Tenant           types.String `tfsdk:"tenant"`
	Region           types.String `tfsdk:"region"`
	ResourceProvider types.String `tfsdk:"resource_provider"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DeletedAt        types.String `tfsdk:"deleted_at"`
	LastModifiedAt   types.String `tfsdk:"last_modified_at"`

	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
	Extensions  types.Map `tfsdk:"extensions"`

	Rules    types.List `tfsdk:"rules"`
	RuleRefs types.List `tfsdk:"rule_refs"`
}

func securityGroupToBaseModel(ctx context.Context, sg *sdk.SecurityGroup) (securityGroupModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := securityGroupModel{}
	model.Id = types.StringValue(sg.Metadata.Ref)
	model.Name = types.StringValue(sg.Metadata.Name)
	model.WorkspaceId = types.StringValue(sg.Metadata.Workspace)
	model.Tenant = types.StringValue(sg.Metadata.Tenant)
	model.Region = types.StringValue(sg.Metadata.Region)
	model.ResourceProvider = refToResourceProvider(sg.Metadata.Ref)
	model.CreatedAt = fromTime(sg.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(sg.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(sg.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, sg.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, sg.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, sg.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	rules, d := sgRulesToListValue(ctx, sg.Spec.Rules)
	diags.Append(d...)
	model.Rules = rules

	ruleRefs, d := sgRuleRefsToListValue(ctx, sg.Spec.RuleRefs)
	diags.Append(d...)
	model.RuleRefs = ruleRefs

	return model, diags
}

type publicIpModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	WorkspaceId      types.String `tfsdk:"workspace_id"`
	Tenant           types.String `tfsdk:"tenant"`
	Region           types.String `tfsdk:"region"`
	ResourceProvider types.String `tfsdk:"resource_provider"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DeletedAt        types.String `tfsdk:"deleted_at"`
	LastModifiedAt   types.String `tfsdk:"last_modified_at"`

	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
	Extensions  types.Map `tfsdk:"extensions"`

	Version    types.String `tfsdk:"version"`
	Address    types.String `tfsdk:"address"`
	AttachedTo types.String `tfsdk:"attached_to"`
	IpAddress  types.String `tfsdk:"ip_address"`
}

func publicIpToBaseModel(ctx context.Context, ip *sdk.PublicIp) (publicIpModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := publicIpModel{}
	model.Id = types.StringValue(ip.Metadata.Ref)
	model.Name = types.StringValue(ip.Metadata.Name)
	model.WorkspaceId = types.StringValue(ip.Metadata.Workspace)
	model.Tenant = types.StringValue(ip.Metadata.Tenant)
	model.Region = types.StringValue(ip.Metadata.Region)
	model.ResourceProvider = refToResourceProvider(ip.Metadata.Ref)
	model.CreatedAt = fromTime(ip.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(ip.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(ip.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, ip.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, ip.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, ip.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	model.Version = types.StringValue(string(ip.Spec.Version))
	if ip.Status != nil {
		model.Address = types.StringValue(ip.Status.IpAddress)
		model.IpAddress = types.StringValue(ip.Status.IpAddress)
		model.AttachedTo = fromRefPtr(ip.Status.AttachedTo)
	} else {
		model.Address = types.StringNull()
		model.IpAddress = types.StringNull()
		model.AttachedTo = types.StringNull()
	}

	return model, diags
}

type subnetModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	WorkspaceId      types.String `tfsdk:"workspace_id"`
	NetworkId        types.String `tfsdk:"network_id"`
	Tenant           types.String `tfsdk:"tenant"`
	Region           types.String `tfsdk:"region"`
	ResourceProvider types.String `tfsdk:"resource_provider"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DeletedAt        types.String `tfsdk:"deleted_at"`
	LastModifiedAt   types.String `tfsdk:"last_modified_at"`

	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
	Extensions  types.Map `tfsdk:"extensions"`

	Cidr         SubnetCidrModel `tfsdk:"cidr"`
	RouteTableId types.String    `tfsdk:"route_table_id"`
	Zone         types.String    `tfsdk:"zone"`
	SkuId        types.String    `tfsdk:"sku_id"`
}

func subnetToBaseModel(ctx context.Context, sub *sdk.Subnet) (subnetModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := subnetModel{}
	model.Id = types.StringValue(sub.Metadata.Ref)
	model.Name = types.StringValue(sub.Metadata.Name)
	model.WorkspaceId = types.StringValue(sub.Metadata.Workspace)
	model.NetworkId = types.StringValue(sub.Metadata.Network)
	model.Tenant = types.StringValue(sub.Metadata.Tenant)
	model.Region = types.StringValue(sub.Metadata.Region)
	model.ResourceProvider = refToResourceProvider(sub.Metadata.Ref)
	model.CreatedAt = fromTime(sub.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(sub.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(sub.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, sub.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, sub.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, sub.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	model.Cidr = SubnetCidrModel{
		Ipv4: types.StringValue(sub.Spec.Cidr.Ipv4),
		Ipv6: types.StringValue(sub.Spec.Cidr.Ipv6),
	}
	model.RouteTableId = types.StringValue(sub.Spec.RouteTableRef.Resource)
	model.Zone = types.StringValue(sub.Spec.Zone)
	model.SkuId = fromRefPtr(sub.Spec.SkuRef)

	return model, diags
}

type routeTableModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	WorkspaceId      types.String `tfsdk:"workspace_id"`
	NetworkId        types.String `tfsdk:"network_id"`
	Tenant           types.String `tfsdk:"tenant"`
	Region           types.String `tfsdk:"region"`
	ResourceProvider types.String `tfsdk:"resource_provider"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DeletedAt        types.String `tfsdk:"deleted_at"`
	LastModifiedAt   types.String `tfsdk:"last_modified_at"`

	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
	Extensions  types.Map `tfsdk:"extensions"`

	Routes types.List `tfsdk:"routes"`
}

func routeTableToBaseModel(ctx context.Context, rt *sdk.RouteTable) (routeTableModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := routeTableModel{}
	model.Id = types.StringValue(rt.Metadata.Ref)
	model.Name = types.StringValue(rt.Metadata.Name)
	model.WorkspaceId = types.StringValue(rt.Metadata.Workspace)
	model.NetworkId = types.StringValue(rt.Metadata.Network)
	model.Tenant = types.StringValue(rt.Metadata.Tenant)
	model.Region = types.StringValue(rt.Metadata.Region)
	model.ResourceProvider = refToResourceProvider(rt.Metadata.Ref)
	model.CreatedAt = fromTime(rt.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(rt.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(rt.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, rt.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, rt.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, rt.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	routes, d := routesToListValue(ctx, rt.Spec.Routes)
	diags.Append(d...)
	model.Routes = routes

	return model, diags
}

type internetGatewayModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	WorkspaceId      types.String `tfsdk:"workspace_id"`
	Tenant           types.String `tfsdk:"tenant"`
	Region           types.String `tfsdk:"region"`
	ResourceProvider types.String `tfsdk:"resource_provider"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DeletedAt        types.String `tfsdk:"deleted_at"`
	LastModifiedAt   types.String `tfsdk:"last_modified_at"`

	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
	Extensions  types.Map `tfsdk:"extensions"`

	EgressOnly types.Bool `tfsdk:"egress_only"`
}

func internetGatewayToBaseModel(ctx context.Context, gtw *sdk.InternetGateway) (internetGatewayModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := internetGatewayModel{}
	model.Id = types.StringValue(gtw.Metadata.Ref)
	model.Name = types.StringValue(gtw.Metadata.Name)
	model.WorkspaceId = types.StringValue(gtw.Metadata.Workspace)
	model.Tenant = types.StringValue(gtw.Metadata.Tenant)
	model.Region = types.StringValue(gtw.Metadata.Region)
	model.ResourceProvider = refToResourceProvider(gtw.Metadata.Ref)
	model.CreatedAt = fromTime(gtw.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(gtw.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(gtw.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, gtw.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, gtw.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, gtw.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	model.EgressOnly = types.BoolValue(gtw.Spec.EgressOnly)

	return model, diags
}
