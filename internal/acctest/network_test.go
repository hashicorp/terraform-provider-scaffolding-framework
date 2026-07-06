package acctest

import (
	"context"
	"fmt"
	"testing"

	"github.com/eu-sovereign-cloud/go-sdk/secapi"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckNetworkDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := testAccRegionalClient(ctx)
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "seca_network" {
			continue
		}

		wref := secapi.WorkspaceReference{
			Tenant:    secapi.TenantID(testAccTenant),
			Workspace: secapi.WorkspaceID(rs.Primary.Attributes["workspace_id"]),
			Name:      rs.Primary.Attributes["name"],
		}

		_, err := client.NetworkV1.GetNetwork(ctx, wref)
		if err == secapi.ErrResourceNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking network %q was destroyed: %w", wref.Name, err)
		}
		return fmt.Errorf("network %q still exists after destroy", wref.Name)
	}

	return nil
}

func testAccNetworkResourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_workspace" "test" {
  name = "workspace-1"
}
resource "seca_network" "test" {
  name         = "network-1"
  workspace_id = seca_workspace.test.name

  sku_id = "N10K"
  cidr = {
    ipv4 = "10.0.0.0/16"
  }
  labels = %s
}
`, formatLabels(labels))
}

func testAccNetworkDataSourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_workspace" "test" {
  name = "workspace-1"
}
resource "seca_network" "test" {
  name         = "network-1"
  workspace_id = seca_workspace.test.name

  sku_id = "N10K"
  cidr = {
    ipv4 = "10.0.0.0/16"
  }
  labels = %s
}
data "seca_network" "test" {
  name         = "network-1"
  workspace_id = seca_workspace.test.name
}`, formatLabels(labels))
}

func TestAccNetwork(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig(map[string]string{"env": "dev"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_network.test", "name", "network-1"),
					resource.TestCheckResourceAttr("seca_network.test", "workspace_id", "workspace-1"),
					resource.TestCheckResourceAttr("seca_network.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("seca_network.test", "region", testAccRegion),
					resource.TestCheckResourceAttr("seca_network.test", "sku_id", "N10K"),
					resource.TestCheckResourceAttr("seca_network.test", "cidr.ipv4", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("seca_network.test", "labels.env", "dev"),
				),
			},
			{
				Config: testAccNetworkResourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_network.test", "name", "network-1"),
					resource.TestCheckResourceAttr("seca_network.test", "labels.env", "prod"),
				),
			},
			{
				ResourceName:      "seca_network.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "workspace-1/network-1",
			},
			{
				Config: testAccNetworkDataSourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_network.test", "name", "network-1"),
					resource.TestCheckResourceAttr("seca_network.test", "workspace_id", "workspace-1"),
					resource.TestCheckResourceAttr("seca_network.test", "sku_id", "N10K"),

					resource.TestCheckResourceAttr("data.seca_network.test", "name", "network-1"),
					resource.TestCheckResourceAttr("data.seca_network.test", "workspace_id", "workspace-1"),
					resource.TestCheckResourceAttr("data.seca_network.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("data.seca_network.test", "region", testAccRegion),
					resource.TestCheckResourceAttr("data.seca_network.test", "sku_id", "N10K"),
					resource.TestCheckResourceAttr("data.seca_network.test", "cidr.ipv4", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("data.seca_network.test", "state", "active"),
				),
			},
		},
	})
}
