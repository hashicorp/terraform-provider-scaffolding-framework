data "seca_workspace" "example" {
  name = "workspace-1"
}

data "seca_security_group" "example" {
  name         = "security-group-1"
  workspace_id = data.seca_workspace.example.id
}

output "security_group_provider" {
  value = data.seca_security_group.example.provider
}
output "security_group_tenant_id" {
  value = data.seca_security_group.example.tenant_id
}
output "security_group_workspace_id" {
  value = data.seca_security_group.example.workspace_id
}

output "security_group_labels" {
  value = data.seca_security_group.example.labels
}
output "security_group_annotations" {
  value = data.seca_security_group.example.annotations
}
output "security_group_extensions" {
  value = data.seca_security_group.example.extensions
}

output "security_group_rules" {
  value = data.seca_security_group.example.rules
}
output "security_group_rule_refs" {
  value = data.seca_security_group.example.rule_refs
}

output "security_group_state" {
  value = data.seca_security_group.example.state
}
