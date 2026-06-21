data "seca_workspace" "example" {
  name = "workspace-1"
}

resource "seca_public_ip" "example" {
  name         = "public-ip-1"
  workspace_id = data.seca_workspace.example.id

  version = "IPv4"
}
