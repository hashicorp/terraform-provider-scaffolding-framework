data "seca_role_assignment" "example" {
  name = "role-assignment-1"
}

output "role_assignment_provider" {
  value = data.seca_role_assignment.example.provider
}
output "role_assignment_tenant_id" {
  value = data.seca_role_assignment.example.tenant_id
}

output "role_assignment_labels" {
  value = data.seca_role_assignment.example.labels
}
output "role_assignment_annotations" {
  value = data.seca_role_assignment.example.annotations
}
output "role_assignment_extensions" {
  value = data.seca_role_assignment.example.extensions
}

output "role_assignment_subs" {
  value = data.seca_role_assignment.example.subs
}
output "role_assignment_scopes" {
  value = data.seca_role_assignment.example.scopes
}
output "role_assignment_roles" {
  value = data.seca_role_assignment.example.roles
}

output "role_assignment_state" {
  value = data.seca_role_assignment.example.state
}
