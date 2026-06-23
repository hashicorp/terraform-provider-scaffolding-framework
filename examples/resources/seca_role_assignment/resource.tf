resource "seca_role_assignment" "example" {
  name = "role-assignment-1"

  subs = ["service-account-1"]
  scopes = [
    {
      tenants    = ["tenant-1"],
      regions    = ["region-1"],
      workspaces = ["workspace-1"]
    }
  ]
  roles = ["role-1"]

  labels      = []
  annotations = []
  extensions  = []
}

output "role_assignment_tenant" {
  value = seca_role_assignment.example.tenant
}
output "role_assignment_resource_provider" {
  value = seca_role_assignment.example.resource_provider
}
output "role_assignment_created_at" {
  value = seca_role_assignment.example.created_at
}
output "role_assignment_deleted_at" {
  value = seca_role_assignment.example.deleted_at
}
output "role_assignment_last_modified_at" {
  value = seca_role_assignment.example.last_modified_at
}

output "role_assignment_subs" {
  value = seca_role_assignment.example.subs
}
output "role_assignment_scopes" {
  value = seca_role_assignment.example.scopes
}
output "role_assignment_roles" {
  value = seca_role_assignment.example.roles
}
