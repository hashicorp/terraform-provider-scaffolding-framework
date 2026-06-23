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

resource "seca_subnet" "example" {
  name         = "subnet-1"
  workspace_id = data.seca_workspace.example.id
  network_id   = data.seca_network.example.id

  route_table_id = data.seca_route_table.example.id
  cidr = {
    ipv4 = "10.100.1.0/24"
  }

  labels      = []
  annotations = []
  extensions  = []
}

output "subnet_tenant" {
  value = seca_subnet.example.tenant
}
output "subnet_workspace_id" {
  value = seca_subnet.example.workspace_id
}
output "subnet_network_id" {
  value = seca_subnet.example.network_id
}
output "subnet_region" {
  value = seca_subnet.example.region
}
output "subnet_resource_provider" {
  value = seca_subnet.example.resource_provider
}
output "subnet_created_at" {
  value = seca_subnet.example.created_at
}
output "subnet_deleted_at" {
  value = seca_subnet.example.deleted_at
}
output "subnet_last_modified_at" {
  value = seca_subnet.example.last_modified_at
}

output "subnet_cidr" {
  value = seca_subnet.example.cidr
}
output "subnet_route_table_id" {
  value = seca_subnet.example.route_table_id
}
output "subnet_sku_id" {
  value = seca_subnet.example.sku_id
}
output "subnet_zone" {
  value = seca_subnet.example.zone
}
