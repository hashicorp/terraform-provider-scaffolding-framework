package acctest

import (
	"context"
	"fmt"
	"testing"

	"github.com/eu-sovereign-cloud/go-sdk/secapi"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckBlockStorageDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := testAccRegionalClient(ctx)
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "seca_block_storage" {
			continue
		}

		wref := secapi.WorkspaceReference{
			Tenant:    secapi.TenantID(testAccTenant),
			Workspace: secapi.WorkspaceID(rs.Primary.Attributes["workspace_id"]),
			Name:      rs.Primary.Attributes["name"],
		}

		_, err := client.StorageV1.GetBlockStorage(ctx, wref)
		if err == secapi.ErrResourceNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking block storage %q was destroyed: %w", wref.Name, err)
		}
		return fmt.Errorf("block storage %q still exists after destroy", wref.Name)
	}

	return nil
}

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
		CheckDestroy:             testAccCheckBlockStorageDestroy,
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
