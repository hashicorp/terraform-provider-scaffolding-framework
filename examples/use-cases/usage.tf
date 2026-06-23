provider "seca" {
  token  = "test-token"
  tenant = "tenant-1"
  region = "region-1"
  global_providers = {
    region_v1        = "http://localhost:3000/providers/seca.region",
    authorization_v1 = "http://localhost:3000/providers/seca.authorization"
  }
}

## Workspace

resource "seca_workspace" "workspace" {
  name = "workspace-1"
}

resource "seca_internet_gateway" "internet_gateway" {
  name         = "internet-gateway-1"
  workspace_id = seca_workspace.workspace.id
}

## Network

data "seca_network_sku" "network_sku" {
  name = "N10K"
}

resource "seca_network" "network" {
  name         = "network-1"
  workspace_id = seca_workspace.workspace.id

  sku_id = data.seca_network_sku.network_sku.id
  cidr = {
    ipv4 = "10.100.0.0/16"
  }
}

resource "seca_route_table" "route_table" {
  name         = "route-table-1"
  workspace_id = seca_workspace.workspace.id
  network_id   = seca_network.network.id

  routes = [
    {
      destination_cidr_block = "0.0.0.0/0"
      target_id              = seca_internet_gateway.internet_gateway.id
    }
  ]
}

resource "seca_subnet" "subnet" {
  name         = "subnet-1"
  workspace_id = seca_workspace.workspace.id
  network_id   = seca_network.network.id

  cidr = {
    ipv4 = "10.100.1.0/24"
  }
  route_table_id = seca_route_table.route_table.id
}

resource "seca_security_group" "security_group" {
  name         = "security-group-1"
  workspace_id = seca_workspace.workspace.id

  rules = [
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        list = [80, 443]
      }
      source_refs = []
    },
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        from = 22
      }
      source_refs = [
        "55.44.33.11"
      ]
    }
  ]
}

resource "seca_public_ip" "public_ip" {
  name         = "public-ip-1"
  workspace_id = seca_workspace.workspace.id

  version = "IPv4"
}

resource "seca_nic" "nic" {
  name         = "nic-1"
  workspace_id = seca_workspace.workspace.id
  subnet_id    = seca_subnet.subnet.id

  addresses    = ["0.0.0.0"]
  public_ip_id = seca_public_ip.public_ip.id
}

# Storage

data "seca_storage_sku" "image_sku" {
  name = "RD100"
}

resource "seca_block_storage" "image_storage" {
  name         = "block-storage-1"
  workspace_id = seca_workspace.workspace.id

  size_gb = 10
  sku_id  = data.seca_storage_sku.image_sku.id
}

resource "seca_image" "image" {
  name = "image-1"

  block_storage_id = seca_block_storage.image_storage.id
  cpu_architecture = "amd64"
  initializer      = "cloudinit-22"
  boot             = "UEFI"
}

data "seca_storage_sku" "storage_sku" {
  name = "RD500"
}

resource "seca_block_storage" "instance_storage" {
  name         = "block-storage-1"
  workspace_id = seca_workspace.workspace.id

  size_gb         = 10
  sku_id          = data.seca_storage_sku.storage_sku.id
  source_image_id = seca_image.image.id
}

# Compute

data "seca_instance_sku" "instance_sku" {
  name = "DXS"
}

resource "seca_instance" "instance" {
  name         = "instance-1"
  workspace_id = seca_workspace.workspace.id

  sku_id         = data.seca_instance_sku.instance_sku.id
  primary_nic_id = seca_nic.nic.id
  zone           = "a"
  ssh_keys       = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl example@secapi.cloud"]

  boot_volume = {
    device_id = seca_block_storage.instance_storage.id
  }
}
