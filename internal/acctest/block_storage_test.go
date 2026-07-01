package acctest

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccBlockStorageResourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_workspace" "test" {
  name = "workspace-1"
}
resource "seca_block_storage" "test" {
  name         = "block-storage-1"
  workspace_id = seca_workspace.test.name

  size_gb = 10
  sku_id  = "storage-skus/RD500"
  labels  = %s
}
`, formatLabels(labels))
}

func testAccBlockStorageDataSourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_workspace" "test" {
  name = "workspace-1"
}
resource "seca_block_storage" "test" {
  name         = "block-storage-1"
  workspace_id = seca_workspace.test.name

  size_gb = 10
  sku_id  = "storage-skus/RD500"
  labels  = %s
}
data "seca_block_storage" "test" {
  name         = "block-storage-1"
  workspace_id = seca_workspace.test.name
}`, formatLabels(labels))
}

func TestAccBlockStorage(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBlockStorageResourceConfig(map[string]string{"env": "dev"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_block_storage.test", "name", "block-storage-1"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "workspace_id", "workspace-1"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "region", "region"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "size_gb", "10"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "labels.env", "dev"),
				),
			},
			{
				Config: testAccBlockStorageResourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_block_storage.test", "name", "block-storage-1"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "labels.env", "prod"),
				),
			},
			{
				ResourceName:      "seca_block_storage.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "workspace-1/block-storage-1",
			},
			{
				Config: testAccBlockStorageDataSourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_block_storage.test", "name", "block-storage-1"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "workspace_id", "workspace-1"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "size_gb", "10"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "labels.env", "prod"),

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
