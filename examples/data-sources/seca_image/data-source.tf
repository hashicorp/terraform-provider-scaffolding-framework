data "seca_image" "example" {
  name = "image-1"
}

output "image_tenant" {
  value = data.seca_image.example.tenant
}
output "image_region" {
  value = data.seca_image.example.region
}
output "image_resource_provider" {
  value = data.seca_image.example.resource_provider
}
output "image_created_at" {
  value = data.seca_image.example.created_at
}
output "image_deleted_at" {
  value = data.seca_image.example.deleted_at
}
output "image_last_modified_at" {
  value = data.seca_image.example.last_modified_at
}

output "image_labels" {
  value = data.seca_image.example.labels
}
output "image_annotations" {
  value = data.seca_image.example.annotations
}
output "image_extensions" {
  value = data.seca_image.example.extensions
}

output "image_block_storage_id" {
  value = data.seca_image.example.block_storage_id
}
output "image_cpu_architecture" {
  value = data.seca_image.example.cpu_architecture
}
output "image_initializer" {
  value = data.seca_image.example.initializer
}
output "image_boot" {
  value = data.seca_image.example.boot
}
output "image_size_mb" {
  value = data.seca_image.example.size_mb
}

output "image_state" {
  value = data.seca_image.example.state
}
