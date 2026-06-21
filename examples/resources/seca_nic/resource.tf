data "seca_workspace" "example" {
  name = "workspace-1"
}

data "seca_subnet" "example" {
  name         = "subnet-1"
  workspace_id = data.seca_workspace.example.id
}

resource "seca_nic" "example" {
  name         = "nic-1"
  workspace_id = data.seca_workspace.example.id
  subnet_id    = data.seca_subnet.example.id

  addresses = ["0.0.0.0"]
}
