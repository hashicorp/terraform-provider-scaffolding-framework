data "seca_workspace" "example" {
  name = "workspace-1"
}

data "seca_network" "example" {
  name         = "network-1"
  workspace_id = data.seca_workspace.example.id
}

output "network_provider" {
  value = data.seca_network.example.provider
}
output "network_tenant_id" {
  value = data.seca_network.example.tenant_id
}
output "network_workspace_id" {
  value = data.seca_network.example.workspace_id
}

output "network_labels" {
  value = data.seca_network.example.labels
}
output "network_annotations" {
  value = data.seca_network.example.annotations
}
output "network_extensions" {
  value = data.seca_network.example.extensions
}

output "network_sku_id" {
  value = data.seca_network.example.sku_id
}
output "network_cidr" {
  value = data.seca_network.example.cidr
}
output "network_additional_cidrs" {
  value = data.seca_network.example.additional_cidrs
}

output "network_state" {
  value = data.seca_network.example.state
}
