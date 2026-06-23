data "seca_role_assignment" "example" {
  name = "role-assignment-1"
}

output "role_assignment_id" {
  value = data.seca_role_assignment.example.id
}
output "role_assignment_tenant" {
  value = data.seca_role_assignment.example.tenant
}
output "role_assignment_resource_provider" {
  value = data.seca_role_assignment.example.resource_provider
}
output "role_assignment_created_at" {
  value = data.seca_role_assignment.example.created_at
}
output "role_assignment_deleted_at" {
  value = data.seca_role_assignment.example.deleted_at
}
output "role_assignment_last_modified_at" {
  value = data.seca_role_assignment.example.last_modified_at
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
