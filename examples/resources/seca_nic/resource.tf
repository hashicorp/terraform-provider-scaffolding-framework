data "seca_workspace" "example" {
  name = "workspace-1"
}

data "seca_subnet" "example" {
  name         = "subnet-1"
  workspace_id = data.seca_workspace.example.id
}

resource "seca_nic" "example" {
  name         = "nic-1"
  workspace_id = data.seca_workspace.example.id
  subnet_id    = data.seca_subnet.example.id

  addresses = ["0.0.0.0"]

  labels      = []
  annotations = []
  extensions  = []
}

output "nic_id" {
  value = seca_nic.example.id
}
output "nic_tenant" {
  value = seca_nic.example.tenant
}
output "nic_workspace_id" {
  value = seca_nic.example.workspace_id
}
output "nic_region" {
  value = seca_nic.example.region
}
output "nic_resource_provider" {
  value = seca_nic.example.resource_provider
}
output "nic_created_at" {
  value = seca_nic.example.created_at
}
output "nic_deleted_at" {
  value = seca_nic.example.deleted_at
}
output "nic_last_modified_at" {
  value = seca_nic.example.last_modified_at
}

output "nic_security_group_ids" {
  value = seca_nic.example.security_group_ids
}
output "nic_addresses" {
  value = seca_nic.example.addresses
}
output "nic_public_ip_ids" {
  value = seca_nic.example.public_ip_ids
}
output "nic_sku_id" {
  value = seca_nic.example.sku_id
}
output "nic_subnet_id" {
  value = seca_nic.example.subnet_id
}
output "nic_mac_address" {
  value = seca_nic.example.mac_address
}
