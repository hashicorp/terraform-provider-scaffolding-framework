data "seca_storage_sku" "example" {
  name = "RD500"
}

data "seca_workspace" "example" {
  name = "workspace-1"
}

resource "seca_block_storage" "example" {
  name         = "block-storage-1"
  workspace_id = data.seca_workspace.example.id

  size_gb = 10
  sku_id  = data.seca_storage_sku.example.id

  labels      = []
  annotations = []
  extensions  = []
}

output "block_storage_tenant" {
  value = seca_block_storage.example.tenant
}
output "block_storage_workspace_id" {
  value = seca_block_storage.example.workspace_id
}
output "block_storage_region" {
  value = seca_block_storage.example.region
}
output "block_storage_resource_provider" {
  value = seca_block_storage.example.resource_provider
}

output "block_storage_sku_id" {
  value = seca_block_storage.example.sku_id
}
output "block_storage_size_gb" {
  value = seca_block_storage.example.size_gb
}
output "block_storage_source_image_id" {
  value = seca_block_storage.example.source_image_id
}
