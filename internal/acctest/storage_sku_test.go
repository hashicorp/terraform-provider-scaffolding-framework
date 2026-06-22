package acctest

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccStorageSkuDataSourceConfig() string {
	return testAccProviderConfig() + `
data "seca_storage_sku" "test" {
  name = "RD500"
}`
}

func TestAccStorageSku(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageSkuDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.seca_storage_sku.test", "name", "RD500"),
					resource.TestCheckResourceAttr("data.seca_storage_sku.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("data.seca_storage_sku.test", "region", "region"),
					resource.TestCheckResourceAttr("data.seca_storage_sku.test", "resource_provider", "seca.storage/v1"),
					resource.TestCheckResourceAttrSet("data.seca_storage_sku.test", "iops"),
					resource.TestCheckResourceAttrSet("data.seca_storage_sku.test", "type"),
					resource.TestCheckResourceAttrSet("data.seca_storage_sku.test", "min_volume_size"),
				),
			},
		},
	})
}
