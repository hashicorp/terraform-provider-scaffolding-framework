package acctest

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccNetworkSkuDataSourceConfig() string {
	return testAccProviderConfig() + `
data "seca_network_sku" "test" {
  name = "N10K"
}`
}

func TestAccNetworkSku(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkSkuDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.seca_network_sku.test", "name", "N10K"),
					resource.TestCheckResourceAttr("data.seca_network_sku.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("data.seca_network_sku.test", "region", testAccRegion),
					resource.TestCheckResourceAttrSet("data.seca_network_sku.test", "bandwidth"),
					resource.TestCheckResourceAttrSet("data.seca_network_sku.test", "packets"),
				),
			},
		},
	})
}
