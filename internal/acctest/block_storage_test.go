package acctest

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccBlockStorageResourceConfig() string {
	return testAccProviderConfig() + `
resource "seca_workspace" "test" {
  name = "workspace-1"
}
resource "seca_block_storage" "test" {
  name         = "block-storage-1"
  workspace_id = seca_workspace.test.name

  size_gb = 10
  sku_id  = "storage-skus/RD500"
}
`
}

func testAccBlockStorageDataSourceConfig() string {
	return testAccProviderConfig() + `
resource "seca_workspace" "test" {
  name = "workspace-1"
}
resource "seca_block_storage" "test" {
  name         = "block-storage-1"
  workspace_id = seca_workspace.test.name

  size_gb = 10
  sku_id  = "storage-skus/RD500"
}
data "seca_block_storage" "test" {
  name         = "block-storage-1"
  workspace_id = seca_workspace.test.name
}`
}

func TestAccBlockStorage(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBlockStorageResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_block_storage.test", "name", "block-storage-1"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "workspace_id", "workspace-1"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "region", "region"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "size_gb", "10"),
				),
			},
			{
				Config: testAccBlockStorageDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_block_storage.test", "name", "block-storage-1"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "workspace_id", "workspace-1"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "size_gb", "10"),

					resource.TestCheckResourceAttr("data.seca_block_storage.test", "name", "block-storage-1"),
					resource.TestCheckResourceAttr("data.seca_block_storage.test", "workspace_id", "workspace-1"),
					resource.TestCheckResourceAttr("data.seca_block_storage.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("data.seca_block_storage.test", "region", "region"),
					resource.TestCheckResourceAttr("data.seca_block_storage.test", "size_gb", "10"),
					resource.TestCheckResourceAttr("data.seca_block_storage.test", "state", "active"),
				),
			},
		},
	})
}
