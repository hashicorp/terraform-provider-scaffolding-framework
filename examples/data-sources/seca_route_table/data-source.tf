data "seca_workspace" "example" {
  name = "workspace-1"
}

data "seca_network" "example" {
  name         = "network-1"
  workspace_id = data.seca_workspace.example.id
}

data "seca_route_table" "example" {
  name         = "route-table-1"
  workspace_id = data.seca_workspace.example.id
  network_id   = data.seca_network.example.id
}

output "route_table_provider" {
  value = data.seca_route_table.example.provider
}
output "route_table_tenant_id" {
  value = data.seca_route_table.example.tenant_id
}
output "route_table_workspace_id" {
  value = data.seca_route_table.example.workspace_id
}
output "route_table_network_id" {
  value = data.seca_route_table.example.network_id
}

output "route_table_labels" {
  value = data.seca_route_table.example.labels
}
output "route_table_annotations" {
  value = data.seca_route_table.example.annotations
}
output "route_table_extensions" {
  value = data.seca_route_table.example.extensions
}

output "route_table_routes" {
  value = data.seca_route_table.example.routes
}

output "route_table_state" {
  value = data.seca_route_table.example.state
}
