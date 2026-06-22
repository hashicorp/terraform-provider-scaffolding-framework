data "seca_workspace" "example" {
  name = "workspace-1"
}

resource "seca_public_ip" "example" {
  name         = "public-ip-1"
  workspace_id = data.seca_workspace.example.id

  version = "IPv4"

  labels      = []
  annotations = []
  extensions  = []
}

output "public_ip_tenant" {
  value = seca_public_ip.example.tenant
}
output "public_ip_workspace_id" {
  value = seca_public_ip.example.workspace_id
}
output "public_ip_region" {
  value = seca_public_ip.example.region
}
output "public_ip_resource_provider" {
  value = seca_public_ip.example.resource_provider
}
output "public_ip_created_at" {
  value = seca_public_ip.example.created_at
}
output "public_ip_deleted_at" {
  value = seca_public_ip.example.deleted_at
}
output "public_ip_last_modified_at" {
  value = seca_public_ip.example.last_modified_at
}

output "public_ip_version" {
  value = seca_public_ip.example.version
}
output "public_ip_address" {
  value = seca_public_ip.example.address
}
output "public_ip_attached_to" {
  value = seca_public_ip.example.attached_to
}
output "public_ip_ip_address" {
  value = seca_public_ip.example.ip_address
}
