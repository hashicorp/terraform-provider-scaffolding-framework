data "seca_instance_sku" "example" {
  name = "DXS"
}

output "instance_sku_provider" {
  value = data.seca_instance_sku.example.provider
}
output "instance_sku_tenant_id" {
  value = data.seca_instance_sku.example.tenant_id
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

output "instance_sku_state" {
  value = data.seca_instance_sku.example.state
}
