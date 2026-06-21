data "seca_storage_sku" "example" {
  name = "RD100"
}

data "seca_workspace" "example" {
  name = "workspace-1"
}

data "seca_block_storage" "example" {
  name         = "block-storage-1"
  workspace_id = data.seca_workspace.example.id
}

resource "seca_image" "example" {
  name = "image-1"

  block_storage_id = data.seca_block_storage.example.id
  cpu_architecture = "amd64"
  initializer      = "cloudinit-22"
  boot             = "UEFI"
}
