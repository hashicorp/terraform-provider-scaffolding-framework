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

output "nic_provider" {
  value = data.seca_nic.example.provider
}
output "nic_tenant_id" {
  value = data.seca_nic.example.tenant_id
}
output "nic_workspace_id" {
  value = data.seca_nic.example.workspace_id
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
