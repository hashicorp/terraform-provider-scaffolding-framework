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
}
