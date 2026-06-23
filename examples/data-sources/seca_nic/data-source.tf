data "seca_workspace" "example" {
  name = "workspace-1"
}

data "seca_subnet" "example" {
  name         = "subnet-1"
  workspace_id = data.seca_workspace.example.id
}

data "seca_nic" "example" {
  name         = "nic-1"
  workspace_id = data.seca_workspace.example.id
  subnet_id    = data.seca_subnet.example.id
}

output "nic_tenant" {
  value = data.seca_nic.example.tenant
}
output "nic_workspace_id" {
  value = data.seca_nic.example.workspace_id
}
output "nic_region" {
  value = data.seca_nic.example.region
}
output "nic_resource_provider" {
  value = data.seca_nic.example.resource_provider
}
output "nic_created_at" {
  value = data.seca_nic.example.created_at
}
output "nic_deleted_at" {
  value = data.seca_nic.example.deleted_at
}
output "nic_last_modified_at" {
  value = data.seca_nic.example.last_modified_at
}

output "nic_labels" {
  value = data.seca_nic.example.labels
}
output "nic_annotations" {
  value = data.seca_nic.example.annotations
}
output "nic_extensions" {
  value = data.seca_nic.example.extensions
}

output "nic_security_group_ids" {
  value = data.seca_nic.example.security_group_ids
}
output "nic_addresses" {
  value = data.seca_nic.example.addresses
}
output "nic_public_ip_ids" {
  value = data.seca_nic.example.public_ip_ids
}
output "nic_sku_id" {
  value = data.seca_nic.example.sku_id
}
output "nic_subnet_id" {
  value = data.seca_nic.example.subnet_id
}
output "nic_mac_address" {
  value = data.seca_nic.example.mac_address
}

output "nic_state" {
  value = data.seca_nic.example.state
}
