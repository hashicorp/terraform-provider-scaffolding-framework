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
}
