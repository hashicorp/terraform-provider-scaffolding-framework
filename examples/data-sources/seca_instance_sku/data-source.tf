data "seca_instance_sku" "example" {
  name = "DXS"
}

output "instance_sku_tenant" {
  value = data.seca_instance_sku.example.tenant
}
output "instance_sku_region" {
  value = data.seca_instance_sku.example.region
}
output "instance_sku_resource_provider" {
  value = data.seca_instance_sku.example.resource_provider
}
output "instance_sku_created_at" {
  value = data.seca_instance_sku.example.created_at
}
output "instance_sku_deleted_at" {
  value = data.seca_instance_sku.example.deleted_at
}
output "instance_sku_last_modified_at" {
  value = data.seca_instance_sku.example.last_modified_at
}

output "instance_sku_labels" {
  value = data.seca_instance_sku.example.labels
}
output "instance_sku_annotations" {
  value = data.seca_instance_sku.example.annotations
}
output "instance_sku_extensions" {
  value = data.seca_instance_sku.example.extensions
}

output "instance_sku_vcpu" {
  value = data.seca_instance_sku.example.vcpu
}
output "instance_sku_ram" {
  value = data.seca_instance_sku.example.ram
}
