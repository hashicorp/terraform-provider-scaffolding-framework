data "seca_workspace" "example" {
  name = "workspace-1"
}

resource "seca_internet_gateway" "example" {
  name         = "internet-gateway-1"
  workspace_id = data.seca_workspace.example.id
}
