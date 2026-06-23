data "seca_role" "example" {
  name = "role-1"
}

output "role_id" {
  value = data.seca_role.example.id
}
output "role_tenant" {
  value = data.seca_role.example.tenant
}
output "role_resource_provider" {
  value = data.seca_role.example.resource_provider
}
output "role_created_at" {
  value = data.seca_role.example.created_at
}
output "role_deleted_at" {
  value = data.seca_role.example.deleted_at
}
output "role_last_modified_at" {
  value = data.seca_role.example.last_modified_at
}

output "role_labels" {
  value = data.seca_role.example.labels
}
output "role_annotations" {
  value = data.seca_role.example.annotations
}
output "role_extensions" {
  value = data.seca_role.example.extensions
}

output "role_permissions" {
  value = data.seca_role.example.permissions
}

output "role_state" {
  value = data.seca_role.example.state
}
