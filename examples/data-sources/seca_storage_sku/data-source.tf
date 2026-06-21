data "seca_storage_sku" "example" {
  name = "RD500"
}

output "storage_sku_provider" {
  value = data.seca_storage_sku.example.provider
}
output "storage_sku_tenant_id" {
  value = data.seca_storage_sku.example.tenant_id
}

output "storage_sku_labels" {
  value = data.seca_storage_sku.example.labels
}
output "storage_sku_annotations" {
  value = data.seca_storage_sku.example.annotations
}
output "storage_sku_extensions" {
  value = data.seca_storage_sku.example.extensions
}

output "storage_sku_iops" {
  value = data.seca_storage_sku.example.iops
}
output "storage_sku_type" {
  value = data.seca_storage_sku.example.type
}
output "storage_sku_min_volume_size" {
  value = data.seca_storage_sku.example.min_volume_size
}

output "storage_sku_state" {
  value = data.seca_storage_sku.example.state
}
