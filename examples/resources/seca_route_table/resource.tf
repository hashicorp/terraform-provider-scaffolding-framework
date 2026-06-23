data "seca_workspace" "example" {
  name = "workspace-1"
}

data "seca_network" "example" {
  name         = "network-1"
  workspace_id = data.seca_workspace.example.id
}

data "seca_internet_gateway" "example" {
  name         = "internet-gateway-1"
  workspace_id = data.seca_workspace.example.id
}

resource "seca_route_table" "example" {
  name         = "route-table-1"
  workspace_id = data.seca_workspace.example.id
  network_id   = data.seca_network.example.id

  routes = [
    {
      destination_cidr_block = "0.0.0.0/0"
      target_id              = data.seca_internet_gateway.example.id
    }
  ]

  labels      = []
  annotations = []
  extensions  = []
}

output "route_table_id" {
  value = seca_route_table.example.id
}
output "route_table_tenant" {
  value = seca_route_table.example.tenant
}
output "route_table_workspace_id" {
  value = seca_route_table.example.workspace_id
}
output "route_table_network_id" {
  value = seca_route_table.example.network_id
}
output "route_table_region" {
  value = seca_route_table.example.region
}
output "route_table_resource_provider" {
  value = seca_route_table.example.resource_provider
}
output "route_table_created_at" {
  value = seca_route_table.example.created_at
}
output "route_table_deleted_at" {
  value = seca_route_table.example.deleted_at
}
output "route_table_last_modified_at" {
  value = seca_route_table.example.last_modified_at
}

output "route_table_routes" {
  value = seca_route_table.example.routes
}
