data "seca_role" "example" {
  name = "role-1"
}

output "role_provider" {
  value = data.seca_role.example.provider
}
output "role_tenant_id" {
  value = data.seca_role.example.tenant_id
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
