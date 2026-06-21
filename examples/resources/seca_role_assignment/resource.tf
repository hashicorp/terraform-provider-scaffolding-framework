resource "seca_role_assignment" "example" {
  name = "role-assignment-1"

  subs = ["service-account-1"]
  scopes = [
    {
      tenants    = ["tenant-1"],
      regions    = ["region-1"],
      workspaces = ["workspace-1"]
    }
  ]
  roles = ["role-1"]
}
