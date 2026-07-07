package acctest

import (
	"context"
	"fmt"
	"testing"

	"github.com/eu-sovereign-cloud/go-sdk/secapi"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckNicDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := testAccRegionalClient(ctx)
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "seca_nic" {
			continue
		}

		wref := secapi.WorkspaceReference{
			Tenant:    secapi.TenantID(testAccTenant),
			Workspace: secapi.WorkspaceID(rs.Primary.Attributes["workspace_id"]),
			Name:      rs.Primary.Attributes["name"],
		}

		_, err := client.NetworkV1.GetNic(ctx, wref)
		if err == secapi.ErrResourceNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking NIC %q was destroyed: %w", wref.Name, err)
		}
		return fmt.Errorf("NIC %q still exists after destroy", wref.Name)
	}

	return nil
}

func testAccNicResourceConfig(labels map[string]string) string {
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
}
resource "seca_subnet" "test" {
  name         = "subnet-1"
  workspace_id = seca_workspace.test.name
  network_id   = seca_network.test.name

  cidr = {
    ipv4 = "10.0.1.0/24"
  }
}
resource "seca_nic" "test" {
  name         = "nic-1"
  workspace_id = seca_workspace.test.name
  subnet_id    = seca_subnet.test.id

  labels = %s
}
`, formatLabels(labels))
}

func testAccNicDataSourceConfig(labels map[string]string) string {
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
}
resource "seca_subnet" "test" {
  name         = "subnet-1"
  workspace_id = seca_workspace.test.name
  network_id   = seca_network.test.name

  cidr = {
    ipv4 = "10.0.1.0/24"
  }
}
resource "seca_nic" "test" {
  name         = "nic-1"
  workspace_id = seca_workspace.test.name
  subnet_id    = seca_subnet.test.id

  labels = %s
}
data "seca_nic" "test" {
  name         = "nic-1"
  workspace_id = seca_workspace.test.name
}`, formatLabels(labels))
}

func TestAccNic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNicResourceConfig(map[string]string{"env": "dev"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_nic.test", "name", "nic-1"),
					resource.TestCheckResourceAttr("seca_nic.test", "workspace_id", "workspace-1"),
					resource.TestCheckResourceAttr("seca_nic.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("seca_nic.test", "region", testAccRegion),
					resource.TestCheckResourceAttrSet("seca_nic.test", "subnet_id"),
					resource.TestCheckResourceAttr("seca_nic.test", "labels.env", "dev"),
				),
			},
			{
				Config: testAccNicResourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_nic.test", "name", "nic-1"),
					resource.TestCheckResourceAttr("seca_nic.test", "labels.env", "prod"),
				),
			},
			{
				ResourceName:      "seca_nic.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "workspace-1/nic-1",
			},
			{
				Config: testAccNicDataSourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_nic.test", "name", "nic-1"),
					resource.TestCheckResourceAttr("seca_nic.test", "workspace_id", "workspace-1"),

					resource.TestCheckResourceAttr("data.seca_nic.test", "name", "nic-1"),
					resource.TestCheckResourceAttr("data.seca_nic.test", "workspace_id", "workspace-1"),
					resource.TestCheckResourceAttr("data.seca_nic.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("data.seca_nic.test", "region", testAccRegion),
					resource.TestCheckResourceAttrSet("data.seca_nic.test", "subnet_id"),
					resource.TestCheckResourceAttr("data.seca_nic.test", "state", "active"),
				),
			},
		},
	})
}
