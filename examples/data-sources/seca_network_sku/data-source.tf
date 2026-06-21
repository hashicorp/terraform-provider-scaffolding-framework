data "seca_network_sku" "example" {
  name = "N10K"
}

output "network_sku_provider" {
  value = data.seca_network_sku.example.provider
}
output "network_sku_tenant_id" {
  value = data.seca_network_sku.example.tenant_id
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

output "network_sku_state" {
  value = data.seca_network_sku.example.state
}
