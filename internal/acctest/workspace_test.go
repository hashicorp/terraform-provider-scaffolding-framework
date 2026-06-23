package acctest

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccWorkspaceResourceConfig() string {
	return testAccProviderConfig() + `
resource "seca_workspace" "test" {
  name = "workspace-1"
}
`
}

func testAccWorkspaceDataSourceConfig() string {
	return testAccProviderConfig() + `
resource "seca_workspace" "test" {
  name = "workspace-1"
}
data "seca_workspace" "test" {
  name = "workspace-1"
}`
}

func TestAccWorkspace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_workspace.test", "name", "workspace-1"),
					resource.TestCheckResourceAttr("seca_workspace.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("seca_workspace.test", "region", "region"),
					resource.TestCheckResourceAttr("seca_workspace.test", "resource_provider", "seca.workspace/v1"),
				),
			},
			{
				Config: testAccWorkspaceDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_workspace.test", "name", "workspace-1"),
					resource.TestCheckResourceAttr("seca_workspace.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("seca_workspace.test", "region", "region"),
					resource.TestCheckResourceAttr("seca_workspace.test", "resource_provider", "seca.workspace/v1"),

					resource.TestCheckResourceAttr("data.seca_workspace.test", "name", "workspace-1"),
					resource.TestCheckResourceAttr("data.seca_workspace.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("seca_workspace.test", "region", "region"),
					resource.TestCheckResourceAttr("data.seca_workspace.test", "resource_provider", "seca.workspace/v1"),
					resource.TestCheckResourceAttr("data.seca_workspace.test", "state", "active"),
				),
			},
		},
	})
}
