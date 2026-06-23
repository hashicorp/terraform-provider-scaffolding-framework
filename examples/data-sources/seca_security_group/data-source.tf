data "seca_workspace" "example" {
  name = "workspace-1"
}

data "seca_security_group" "example" {
  name         = "security-group-1"
  workspace_id = data.seca_workspace.example.id
}

output "security_group_id" {
  value = data.seca_security_group.example.id
}
output "security_group_tenant" {
  value = data.seca_security_group.example.tenant
}
output "security_group_workspace_id" {
  value = data.seca_security_group.example.workspace_id
}
output "security_group_region" {
  value = data.seca_security_group.example.region
}
output "security_group_resource_provider" {
  value = data.seca_security_group.example.resource_provider
}
output "security_group_created_at" {
  value = data.seca_security_group.example.created_at
}
output "security_group_deleted_at" {
  value = data.seca_security_group.example.deleted_at
}
output "security_group_last_modified_at" {
  value = data.seca_security_group.example.last_modified_at
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
