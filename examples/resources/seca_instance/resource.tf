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
}
