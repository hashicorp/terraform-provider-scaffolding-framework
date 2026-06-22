package acctest

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccImageResourceConfig() string {
	return testAccProviderConfig() + `
resource "seca_image" "test" {
  name = "image-1"

  block_storage_id = "block-storages/block-storage-1"
  cpu_architecture = "amd64"
  initializer      = "cloudinit-22"
  boot             = "UEFI"
}
`
}

func testAccImageDataSourceConfig() string {
	return testAccProviderConfig() + `
resource "seca_image" "test" {
  name = "image-1"

  block_storage_id = "block-storages/block-storage-1"
  cpu_architecture = "amd64"
  initializer      = "cloudinit-22"
  boot             = "UEFI"
}
data "seca_image" "test" {
  name = "image-1"
}`
}

func TestAccImage(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccImageResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_image.test", "name", "image-1"),
					resource.TestCheckResourceAttr("seca_image.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("seca_image.test", "region", "region"),
					resource.TestCheckResourceAttr("seca_image.test", "resource_provider", "seca.storage/v1"),
					resource.TestCheckResourceAttr("seca_image.test", "cpu_architecture", "amd64"),
					resource.TestCheckResourceAttr("seca_image.test", "initializer", "cloudinit-22"),
					resource.TestCheckResourceAttr("seca_image.test", "boot", "UEFI"),
				),
			},
			{
				Config: testAccImageDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_image.test", "name", "image-1"),
					resource.TestCheckResourceAttr("seca_image.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("seca_image.test", "region", "region"),
					resource.TestCheckResourceAttr("seca_image.test", "resource_provider", "seca.storage/v1"),

					resource.TestCheckResourceAttr("data.seca_image.test", "name", "image-1"),
					resource.TestCheckResourceAttr("data.seca_image.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("data.seca_image.test", "region", "region"),
					resource.TestCheckResourceAttr("data.seca_image.test", "resource_provider", "seca.storage/v1"),
					resource.TestCheckResourceAttr("data.seca_image.test", "cpu_architecture", "amd64"),
					resource.TestCheckResourceAttr("data.seca_image.test", "initializer", "cloudinit-22"),
					resource.TestCheckResourceAttr("data.seca_image.test", "boot", "UEFI"),
					resource.TestCheckResourceAttr("data.seca_image.test", "state", "active"),
				),
			},
		},
	})
}
