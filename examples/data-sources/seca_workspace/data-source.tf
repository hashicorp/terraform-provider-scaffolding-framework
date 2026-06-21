data "seca_workspace" "example" {
  name   = "workspace-1"
}

output "workspace_tenant" {
  value = data.seca_workspace.example.tenant
}
output "workspace_resource_provider" {
  value = data.seca_workspace.example.provider
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
