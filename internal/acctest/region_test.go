package acctest

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccRegionDataSourceConfig() string {
	return testAccProviderConfig() + `
data "seca_region" "test" {
  name = "region"
}`
}

func TestAccRegion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.seca_region.test", "name", "region"),
					resource.TestCheckResourceAttrSet("data.seca_region.test", "available_zones.#"),
					resource.TestCheckResourceAttrSet("data.seca_region.test", "providers.#"),
				),
			},
		},
	})
}
