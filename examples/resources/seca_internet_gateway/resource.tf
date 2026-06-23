data "seca_workspace" "example" {
  name = "workspace-1"
}

resource "seca_internet_gateway" "example" {
  name         = "internet-gateway-1"
  workspace_id = data.seca_workspace.example.id

  labels      = []
  annotations = []
  extensions  = []
}

output "internet_gateway_tenant" {
  value = seca_internet_gateway.example.tenant
}
output "internet_gateway_workspace_id" {
  value = seca_internet_gateway.example.workspace_id
}
output "internet_gateway_region" {
  value = seca_internet_gateway.example.region
}
output "internet_gateway_resource_provider" {
  value = seca_internet_gateway.example.resource_provider
}
output "internet_gateway_created_at" {
  value = seca_internet_gateway.example.created_at
}
output "internet_gateway_deleted_at" {
  value = seca_internet_gateway.example.deleted_at
}
output "internet_gateway_last_modified_at" {
  value = seca_internet_gateway.example.last_modified_at
}

output "internet_gateway_egress_only" {
  value = seca_internet_gateway.example.egress_only
}
