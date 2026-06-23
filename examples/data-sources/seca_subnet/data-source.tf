data "seca_workspace" "example" {
  name = "workspace-1"
}

data "seca_network" "example" {
  name         = "network-1"
  workspace_id = data.seca_workspace.example.id
}

data "seca_subnet" "example" {
  name         = "subnet-1"
  workspace_id = data.seca_workspace.example.id
  network_id   = data.seca_network.example.id
}

output "subnet_id" {
  value = data.seca_subnet.example.id
}
output "subnet_tenant" {
  value = data.seca_subnet.example.tenant
}
output "subnet_workspace_id" {
  value = data.seca_subnet.example.workspace_id
}
output "route_table_network_id" {
  value = data.seca_route_table.example.network_id
}
output "subnet_region" {
  value = data.seca_subnet.example.region
}
output "subnet_resource_provider" {
  value = data.seca_subnet.example.resource_provider
}
output "subnet_created_at" {
  value = data.seca_subnet.example.created_at
}
output "subnet_deleted_at" {
  value = data.seca_subnet.example.deleted_at
}
output "subnet_last_modified_at" {
  value = data.seca_subnet.example.last_modified_at
}

output "subnet_labels" {
  value = data.seca_subnet.example.labels
}
output "subnet_annotations" {
  value = data.seca_subnet.example.annotations
}
output "subnet_extensions" {
  value = data.seca_subnet.example.extensions
}

output "subnet_cidr" {
  value = data.seca_subnet.example.cidr
}
output "subnet_route_table_id" {
  value = data.seca_subnet.example.route_table_id
}
output "subnet_sku_id" {
  value = data.seca_subnet.example.sku_id
}
output "subnet_zone" {
  value = data.seca_subnet.example.zone
}

output "subnet_state" {
  value = data.seca_subnet.example.state
}
