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
}
