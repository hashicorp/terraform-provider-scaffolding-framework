data "seca_workspace" "example" {
  name = "workspace-1"
}

data "seca_instance" "example" {
  name         = "instance-1"
  workspace_id = data.seca_workspace.example.id
}

output "instance_id" {
  value = data.seca_instance.example.id
}
output "instance_tenant" {
  value = data.seca_instance.example.tenant
}
output "instance_workspace_id" {
  value = data.seca_instance.example.workspace_id
}
output "instance_region" {
  value = data.seca_instance.example.region
}
output "instance_resource_provider" {
  value = data.seca_instance.example.resource_provider
}
output "instance_created_at" {
  value = data.seca_instance.example.created_at
}
output "instance_deleted_at" {
  value = data.seca_instance.example.deleted_at
}
output "instance_last_modified_at" {
  value = data.seca_instance.example.last_modified_at
}

output "instance_labels" {
  value = data.seca_instance.example.labels
}
output "instance_annotations" {
  value = data.seca_instance.example.annotations
}
output "instance_extensions" {
  value = data.seca_instance.example.extensions
}

output "instance_sku_id" {
  value = data.seca_instance.example.sku_id
}
output "instance_primary_nic_id" {
  value = data.seca_instance.example.primary_nic_id
}
output "instance_additional_nic_ids" {
  value = data.seca_instance.example.additional_nic_ids
}
output "instance_zone" {
  value = data.seca_instance.example.zone
}
output "instance_security_group_id" {
  value = data.seca_instance.example.security_group_id
}
output "instance_user_data" {
  value = data.seca_instance.example.user_data
}
output "instance_anti_affinity_group" {
  value = data.seca_instance.example.anti_affinity_group
}
output "instance_ssh_keys" {
  value = data.seca_instance.example.ssh_keys
}
output "instance_boot_volume" {
  value = data.seca_instance.example.boot_volume
}
output "instance_data_volumes" {
  value = data.seca_instance.example.data_volumes
}
output "instance_power_state" {
  value = data.seca_instance.example.power_state
}
output "instance_power_state_since" {
  value = data.seca_instance.example.power_state_since
}

output "instance_state" {
  value = data.seca_instance.example.state
}
