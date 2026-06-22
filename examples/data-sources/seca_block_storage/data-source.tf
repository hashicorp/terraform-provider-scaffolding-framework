data "seca_workspace" "example" {
  name = "workspace-1"
}

data "seca_block_storage" "example" {
  name         = "block-storage-1"
  workspace_id = data.seca_workspace.example.id
}

output "block_storage_tenant" {
  value = data.seca_block_storage.example.tenant
}
output "block_storage_workspace_id" {
  value = data.seca_block_storage.example.workspace_id
}
output "block_storage_region" {
  value = data.seca_block_storage.example.region
}
output "block_storage_resource_provider" {
  value = data.seca_block_storage.example.resource_provider
}
output "block_storage_created_at" {
  value = data.seca_block_storage.example.created_at
}
output "block_storage_deleted_at" {
  value = data.seca_block_storage.example.deleted_at
}
output "block_storage_last_modified_at" {
  value = data.seca_block_storage.example.last_modified_at
}

output "block_storage_labels" {
  value = data.seca_block_storage.example.labels
}
output "block_storage_annotations" {
  value = data.seca_block_storage.example.annotations
}
output "block_storage_extensions" {
  value = data.seca_block_storage.example.extensions
}

output "block_storage_sku_id" {
  value = data.seca_block_storage.example.sku_id
}
output "block_storage_size_gb" {
  value = data.seca_block_storage.example.size_gb
}
output "block_storage_source_image_id" {
  value = data.seca_block_storage.example.source_image_id
}

output "block_storage_state" {
  value = data.seca_block_storage.example.state
}
