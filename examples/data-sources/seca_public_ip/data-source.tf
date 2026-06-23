data "seca_workspace" "example" {
  name = "workspace-1"
}

data "seca_public_ip" "example" {
  name         = "public-ip-1"
  workspace_id = data.seca_workspace.example.id
}

output "public_ip_tenant" {
  value = data.seca_public_ip.example.tenant
}
output "public_ip_workspace_id" {
  value = data.seca_public_ip.example.workspace_id
}
output "public_ip_region" {
  value = data.seca_public_ip.example.region
}
output "public_ip_resource_provider" {
  value = data.seca_public_ip.example.resource_provider
}
output "public_ip_created_at" {
  value = data.seca_public_ip.example.created_at
}
output "public_ip_deleted_at" {
  value = data.seca_public_ip.example.deleted_at
}
output "public_ip_last_modified_at" {
  value = data.seca_public_ip.example.last_modified_at
}

output "public_ip_labels" {
  value = data.seca_public_ip.example.labels
}
output "public_ip_annotations" {
  value = data.seca_public_ip.example.annotations
}
output "public_ip_extensions" {
  value = data.seca_public_ip.example.extensions
}

output "public_ip_version" {
  value = data.seca_public_ip.example.version
}
output "public_ip_address" {
  value = data.seca_public_ip.example.address
}
output "public_ip_attached_to" {
  value = data.seca_public_ip.example.attached_to
}
output "public_ip_ip_address" {
  value = data.seca_public_ip.example.ip_address
}

output "public_ip_state" {
  value = data.seca_public_ip.example.state
}
