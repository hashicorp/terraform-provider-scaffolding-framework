provider "seca" {
  token  = "test-token"
  tenant = "tenant-1"
  region = "region-1"
  retry = {
    delay        = 30 # seconds
    interval     = 10 # seconds
    max_attempts = 5
  }
  global_providers = {
    region_v1        = "http://localhost:3000/providers/seca.region",
    authorization_v1 = "http://localhost:3000/providers/seca.authorization"
  }
}
