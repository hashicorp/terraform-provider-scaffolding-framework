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
}
