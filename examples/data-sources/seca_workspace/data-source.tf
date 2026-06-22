data "seca_workspace" "example" {
  name = "workspace-1"
}

output "workspace_tenant" {
  value = data.seca_workspace.example.tenant
}
output "workspace_resource_region" {
  value = data.seca_workspace.example.region
}
output "workspace_resource_provider" {
  value = data.seca_workspace.example.resource_provider
}
output "workspace_resource_created_at" {
  value = data.seca_workspace.example.created_at
}
output "workspace_resource_deleted_at" {
  value = data.seca_workspace.example.deleted_at
}
output "workspace_resource_last_modified_at" {
  value = data.seca_workspace.example.last_modified_at
}

output "workspace_labels" {
  value = data.seca_workspace.example.labels
}
output "workspace_annotations" {
  value = data.seca_workspace.example.annotations
}
output "workspace_extensions" {
  value = data.seca_workspace.example.extensions
}

output "workspace_state" {
  value = data.seca_workspace.example.state
}
