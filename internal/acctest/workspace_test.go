package acctest

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccWorkspaceResourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_workspace" "test" {
  name   = "workspace-1"
  labels = %s
}
`, formatLabels(labels))
}

func testAccWorkspaceDataSourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_workspace" "test" {
  name   = "workspace-1"
  labels = %s
}
data "seca_workspace" "test" {
  name = "workspace-1"
}`, formatLabels(labels))
}

func TestAccWorkspace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceResourceConfig(map[string]string{"env": "dev"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_workspace.test", "name", "workspace-1"),
					resource.TestCheckResourceAttr("seca_workspace.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("seca_workspace.test", "region", "region"),
					resource.TestCheckResourceAttr("seca_workspace.test", "labels.env", "dev"),
				),
			},
			{
				Config: testAccWorkspaceResourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_workspace.test", "name", "workspace-1"),
					resource.TestCheckResourceAttr("seca_workspace.test", "labels.env", "prod"),
				),
			},
			{
				Config: testAccWorkspaceDataSourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_workspace.test", "name", "workspace-1"),
					resource.TestCheckResourceAttr("seca_workspace.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("seca_workspace.test", "region", "region"),

					resource.TestCheckResourceAttr("data.seca_workspace.test", "name", "workspace-1"),
					resource.TestCheckResourceAttr("data.seca_workspace.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("seca_workspace.test", "region", "region"),
					resource.TestCheckResourceAttr("data.seca_workspace.test", "state", "active"),
				),
			},
		},
	})
}
