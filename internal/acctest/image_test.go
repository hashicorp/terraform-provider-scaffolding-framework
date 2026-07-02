package acctest

import (
	"context"
	"fmt"
	"testing"

	"github.com/eu-sovereign-cloud/go-sdk/secapi"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckImageDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := testAccRegionalClient(ctx)
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "seca_image" {
			continue
		}

		tref := secapi.TenantReference{
			Tenant: secapi.TenantID(testAccTenant),
			Name:   rs.Primary.Attributes["name"],
		}

		_, err := client.StorageV1.GetImage(ctx, tref)
		if err == secapi.ErrResourceNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking image %q was destroyed: %w", tref.Name, err)
		}
		return fmt.Errorf("image %q still exists after destroy", tref.Name)
	}

	return nil
}

func testAccImageResourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_image" "test" {
  name = "image-1"

  block_storage_id = "block-storages/block-storage-1"
  cpu_architecture = "amd64"
  initializer      = "cloudinit-22"
  boot             = "UEFI"
  labels           = %s
}
`, formatLabels(labels))
}

func testAccImageDataSourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_image" "test" {
  name = "image-1"

  block_storage_id = "block-storages/block-storage-1"
  cpu_architecture = "amd64"
  initializer      = "cloudinit-22"
  boot             = "UEFI"
  labels           = %s
}
data "seca_image" "test" {
  name = "image-1"
}`, formatLabels(labels))
}

func TestAccImage(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageResourceConfig(map[string]string{"env": "dev"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_image.test", "name", "image-1"),
					resource.TestCheckResourceAttr("seca_image.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("seca_image.test", "region", "region"),
					resource.TestCheckResourceAttr("seca_image.test", "cpu_architecture", "amd64"),
					resource.TestCheckResourceAttr("seca_image.test", "initializer", "cloudinit-22"),
					resource.TestCheckResourceAttr("seca_image.test", "boot", "UEFI"),
					resource.TestCheckResourceAttr("seca_image.test", "labels.env", "dev"),
				),
			},
			{
				Config: testAccImageResourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_image.test", "name", "image-1"),
					resource.TestCheckResourceAttr("seca_image.test", "labels.env", "prod"),
				),
			},
			{
				ResourceName:      "seca_image.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "image-1",
			},
			{
				Config: testAccImageDataSourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_image.test", "name", "image-1"),
					resource.TestCheckResourceAttr("seca_image.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("seca_image.test", "region", "region"),

					resource.TestCheckResourceAttr("data.seca_image.test", "name", "image-1"),
					resource.TestCheckResourceAttr("data.seca_image.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("data.seca_image.test", "region", "region"),
					resource.TestCheckResourceAttr("data.seca_image.test", "cpu_architecture", "amd64"),
					resource.TestCheckResourceAttr("data.seca_image.test", "initializer", "cloudinit-22"),
					resource.TestCheckResourceAttr("data.seca_image.test", "boot", "UEFI"),
					resource.TestCheckResourceAttr("data.seca_image.test", "state", "active"),
				),
			},
		},
	})
}
