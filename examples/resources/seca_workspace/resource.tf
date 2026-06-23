resource "seca_workspace" "example" {
  name = "workspace-1"

  labels      = []
  annotations = []
  extensions  = []
}

output "workspace_id" {
  value = seca_workspace.example.id
}
output "workspace_tenant" {
  value = seca_workspace.example.tenant
}
output "workspace_resource_region" {
  value = seca_workspace.example.region
}
output "workspace_resource_created_at" {
  value = seca_workspace.example.created_at
}
output "workspace_resource_deleted_at" {
  value = seca_workspace.example.deleted_at
}
output "workspace_resource_last_modified_at" {
  value = seca_workspace.example.last_modified_at
}
