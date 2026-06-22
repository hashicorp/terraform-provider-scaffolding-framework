resource "seca_role" "example" {
  name = "role-1"

  permissions = [
    {
      provider = "seca.network/v1",
      resources = [
        "networks/*",
        "subnets/*",
        "route-tables/*",
        "nics/*",
        "internet-gateways/*",
        "security-groups/*",
        "public-ips/*"
      ],
      verb = ["get", "list"]
    }
  ]

  labels      = []
  annotations = []
  extensions  = []
}

output "role_tenant" {
  value = seca_role.example.tenant
}
output "role_resource_provider" {
  value = seca_role.example.resource_provider
}
output "role_created_at" {
  value = seca_role.example.created_at
}
output "role_deleted_at" {
  value = seca_role.example.deleted_at
}
output "role_last_modified_at" {
  value = seca_role.example.last_modified_at
}

output "role_permissions" {
  value = seca_role.example.permissions
}
