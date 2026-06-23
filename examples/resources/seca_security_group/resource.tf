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

  labels      = []
  annotations = []
  extensions  = []
}

output "security_group_id" {
  value = seca_security_group.example.id
}
output "security_group_tenant" {
  value = seca_security_group.example.tenant
}
output "security_group_workspace_id" {
  value = seca_security_group.example.workspace_id
}
output "security_group_region" {
  value = seca_security_group.example.region
}
output "security_group_resource_provider" {
  value = seca_security_group.example.resource_provider
}
output "security_group_created_at" {
  value = seca_security_group.example.created_at
}
output "security_group_deleted_at" {
  value = seca_security_group.example.deleted_at
}
output "security_group_last_modified_at" {
  value = seca_security_group.example.last_modified_at
}

output "security_group_rules" {
  value = seca_security_group.example.rules
}
output "security_group_rule_refs" {
  value = seca_security_group.example.rule_refs
}
