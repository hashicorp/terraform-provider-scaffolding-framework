# A block storage is workspace-scoped, so it is imported by a composite
# "<workspace_id>/<name>" identifier (the tenant comes from the provider configuration).
import {
  to = seca_block_storage.example
  id = "workspace-1/block-storage-1"
}
