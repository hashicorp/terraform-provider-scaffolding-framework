data "seca_network_sku" "example" {
  name = "N10K"
}

data "seca_workspace" "example" {
  name = "workspace-1"
}

resource "seca_network" "example" {
  name         = "network-1"
  workspace_id = data.seca_workspace.example.id

  sku_id = data.seca_network_sku.example.id
  cidr = {
    ipv4 = "10.100.0.0/16"
  }

  labels      = []
  annotations = []
  extensions  = []
}

output "network_tenant" {
  value = seca_network.example.tenant
}
output "network_workspace_id" {
  value = seca_network.example.workspace_id
}
output "network_region" {
  value = seca_network.example.region
}
output "network_resource_provider" {
  value = seca_network.example.resource_provider
}
output "network_created_at" {
  value = seca_network.example.created_at
}
output "network_deleted_at" {
  value = seca_network.example.deleted_at
}
output "network_last_modified_at" {
  value = seca_network.example.last_modified_at
}

output "network_sku_id" {
  value = seca_network.example.sku_id
}
output "network_cidr" {
  value = seca_network.example.cidr
}
output "network_additional_cidrs" {
  value = seca_network.example.additional_cidrs
}
