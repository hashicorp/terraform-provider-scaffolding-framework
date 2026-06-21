data "seca_workspace" "example" {
  name = "workspace-1"
}

data "seca_internet_gateway" "example" {
  name         = "internet-gateway-1"
  workspace_id = data.seca_workspace.example.id
}

output "internet_gateway_provider" {
  value = data.seca_internet_gateway.example.provider
}
output "internet_gateway_tenant_id" {
  value = data.seca_internet_gateway.example.tenant_id
}
output "internet_gateway_workspace_id" {
  value = data.seca_internet_gateway.example.workspace_id
}

output "internet_gateway_labels" {
  value = data.seca_internet_gateway.example.labels
}
output "internet_gateway_annotations" {
  value = data.seca_internet_gateway.example.annotations
}
output "internet_gateway_extensions" {
  value = data.seca_internet_gateway.example.extensions
}

output "internet_gateway_egress_only" {
  value = data.seca_internet_gateway.example.egress_only
}

output "internet_gateway_state" {
  value = data.seca_internet_gateway.example.state
}
