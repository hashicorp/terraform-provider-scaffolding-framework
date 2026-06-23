data "seca_network_sku" "example" {
  name = "N10K"
}

output "network_sku_id" {
  value = data.seca_network_sku.example.id
}
output "network_sku_tenant" {
  value = data.seca_network_sku.example.tenant
}
output "network_sku_region" {
  value = data.seca_network_sku.example.region
}
output "network_sku_resource_provider" {
  value = data.seca_network_sku.example.resource_provider
}

output "network_sku_labels" {
  value = data.seca_network_sku.example.labels
}
output "network_sku_annotations" {
  value = data.seca_network_sku.example.annotations
}
output "network_sku_extensions" {
  value = data.seca_network_sku.example.extensions
}

output "network_sku_bandwidth" {
  value = data.seca_network_sku.example.bandwidth
}
output "network_sku_packets" {
  value = data.seca_network_sku.example.packets
}
