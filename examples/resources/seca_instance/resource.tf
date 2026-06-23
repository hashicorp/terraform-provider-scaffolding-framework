data "seca_instance_sku" "example" {
  name = "DXS"
}

data "seca_workspace" "example" {
  name = "workspace-1"
}

data "seca_nic" "example" {
  name         = "nic-1"
  workspace_id = data.seca_workspace.example.id
}

data "seca_block_storage" "example" {
  name         = "block-storage-1"
  workspace_id = data.seca_workspace.example.id
}

resource "seca_instance" "example" {
  name         = "instance-1"
  workspace_id = data.seca_workspace.example.id

  sku_id         = data.seca_instance_sku.example.id
  primary_nic_id = data.seca_nic.example.id
  ssh_keys       = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl example@secapi.cloud"]

  boot_volume = {
    device_id = data.seca_block_storage.example.id
  }

  labels      = []
  annotations = []
  extensions  = []
}

output "instance_tenant" {
  value = seca_instance.example.tenant
}
output "instance_workspace_id" {
  value = seca_instance.example.workspace_id
}
output "instance_region" {
  value = seca_instance.example.region
}
output "instance_resource_provider" {
  value = seca_instance.example.resource_provider
}
output "instance_created_at" {
  value = seca_instance.example.created_at
}
output "instance_deleted_at" {
  value = seca_instance.example.deleted_at
}
output "instance_last_modified_at" {
  value = seca_instance.example.last_modified_at
}

output "instance_sku_id" {
  value = seca_instance.example.sku_id
}
output "instance_primary_nic_id" {
  value = seca_instance.example.primary_nic_id
}
output "instance_additional_nic_ids" {
  value = seca_instance.example.additional_nic_ids
}
output "instance_zone" {
  value = seca_instance.example.zone
}
output "instance_security_group_id" {
  value = seca_instance.example.security_group_id
}
output "instance_user_data" {
  value = seca_instance.example.user_data
}
output "instance_anti_affinity_group" {
  value = seca_instance.example.anti_affinity_group
}
output "instance_ssh_keys" {
  value = seca_instance.example.ssh_keys
}
output "instance_boot_volume" {
  value = seca_instance.example.boot_volume
}
output "instance_data_volumes" {
  value = seca_instance.example.data_volumes
}
output "instance_power_state" {
  value = seca_instance.example.power_state
}
output "instance_power_state_since" {
  value = seca_instance.example.power_state_since
}
