data "seca_image" "example" {
  name = "image-1"
}

output "image_provider" {
  value = data.seca_image.example.provider
}
output "image_tenant_id" {
  value = data.seca_image.example.tenant_id
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
