data "seca_region" "example" {
  name = "region-1"
}

output "region_resource_provider" {
  value = data.seca_region.example.resource_provider
}
output "region_created_at" {
  value = data.seca_region.example.created_at
}
output "region_deleted_at" {
  value = data.seca_region.example.deleted_at
}
output "region_last_modified_at" {
  value = data.seca_region.example.last_modified_at
}

output "region_available_zones" {
  value = data.seca_region.example.available_zones
}
output "region_providers" {
  value = data.seca_region.example.providers
}
