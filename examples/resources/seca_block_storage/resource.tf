data "seca_storage_sku" "example" {
  name = "RD500"
}

data "seca_workspace" "example" {
  name = "workspace-1"
}

resource "seca_block_storage" "example" {
  name         = "block-storage-1"
  workspace_id = data.seca_workspace.example.id

  size_gb = 10
  sku_id  = data.seca_storage_sku.example.id
}
