data "seca_storage_sku" "example" {
  name = "RD100"
}

data "seca_workspace" "example" {
  name = "workspace-1"
}

data "seca_block_storage" "example" {
  name         = "block-storage-1"
  workspace_id = data.seca_workspace.example.id
}

resource "seca_image" "example" {
  name = "image-1"

  block_storage_id = data.seca_block_storage.example.id
  cpu_architecture = "amd64"
  initializer      = "cloudinit-22"
  boot             = "UEFI"

  labels      = []
  annotations = []
  extensions  = []
}

output "image_tenant" {
  value = seca_image.example.tenant
}
output "image_region" {
  value = seca_image.example.region
}
output "image_resource_provider" {
  value = seca_image.example.resource_provider
}
output "image_created_at" {
  value = seca_image.example.created_at
}
output "image_deleted_at" {
  value = seca_image.example.deleted_at
}
output "image_last_modified_at" {
  value = seca_image.example.last_modified_at
}

output "image_block_storage_id" {
  value = seca_image.example.block_storage_id
}
output "image_cpu_architecture" {
  value = seca_image.example.cpu_architecture
}
output "image_initializer" {
  value = seca_image.example.initializer
}
output "image_boot" {
  value = seca_image.example.boot
}
