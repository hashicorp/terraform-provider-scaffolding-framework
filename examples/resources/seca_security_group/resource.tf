data "seca_workspace" "example" {
  name = "workspace-1"
}

resource "seca_security_group" "example" {
  name         = "security-group-1"
  workspace_id = data.seca_workspace.example.id

  rules = [
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        list = [80, 443]
      }
      source_refs = []
    },
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        from = 22
      }
      source_refs = [
        "55.44.33.11"
      ]
    }
  ]
}
