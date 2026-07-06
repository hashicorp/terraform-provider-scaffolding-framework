package acctest

import (
	"context"
	"fmt"
	"testing"

	"github.com/eu-sovereign-cloud/go-sdk/secapi"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckPublicIpDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := testAccRegionalClient(ctx)
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "seca_public_ip" {
			continue
		}

		wref := secapi.WorkspaceReference{
			Tenant:    secapi.TenantID(testAccTenant),
			Workspace: secapi.WorkspaceID(rs.Primary.Attributes["workspace_id"]),
			Name:      rs.Primary.Attributes["name"],
		}

		_, err := client.NetworkV1.GetPublicIp(ctx, wref)
		if err == secapi.ErrResourceNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking public IP %q was destroyed: %w", wref.Name, err)
		}
		return fmt.Errorf("public IP %q still exists after destroy", wref.Name)
	}

	return nil
}

func testAccPublicIpResourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_workspace" "test" {
  name = "workspace-1"
}
resource "seca_public_ip" "test" {
  name         = "public-ip-1"
  workspace_id = seca_workspace.test.name

  version = "IPv4"
  labels  = %s
}
`, formatLabels(labels))
}

func testAccPublicIpDataSourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_workspace" "test" {
  name = "workspace-1"
}
resource "seca_public_ip" "test" {
  name         = "public-ip-1"
  workspace_id = seca_workspace.test.name

  version = "IPv4"
  labels  = %s
}
data "seca_public_ip" "test" {
  name         = "public-ip-1"
  workspace_id = seca_workspace.test.name
}`, formatLabels(labels))
}

func TestAccPublicIp(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPublicIpDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPublicIpResourceConfig(map[string]string{"env": "dev"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_public_ip.test", "name", "public-ip-1"),
					resource.TestCheckResourceAttr("seca_public_ip.test", "workspace_id", "workspace-1"),
					resource.TestCheckResourceAttr("seca_public_ip.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("seca_public_ip.test", "region", testAccRegion),
					resource.TestCheckResourceAttr("seca_public_ip.test", "version", "IPv4"),
					resource.TestCheckResourceAttr("seca_public_ip.test", "labels.env", "dev"),
				),
			},
			{
				Config: testAccPublicIpResourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_public_ip.test", "name", "public-ip-1"),
					resource.TestCheckResourceAttr("seca_public_ip.test", "labels.env", "prod"),
				),
			},
			{
				ResourceName:      "seca_public_ip.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "workspace-1/public-ip-1",
			},
			{
				Config: testAccPublicIpDataSourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_public_ip.test", "name", "public-ip-1"),
					resource.TestCheckResourceAttr("seca_public_ip.test", "workspace_id", "workspace-1"),
					resource.TestCheckResourceAttr("seca_public_ip.test", "version", "IPv4"),

					resource.TestCheckResourceAttr("data.seca_public_ip.test", "name", "public-ip-1"),
					resource.TestCheckResourceAttr("data.seca_public_ip.test", "workspace_id", "workspace-1"),
					resource.TestCheckResourceAttr("data.seca_public_ip.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("data.seca_public_ip.test", "region", testAccRegion),
					resource.TestCheckResourceAttr("data.seca_public_ip.test", "version", "IPv4"),
					resource.TestCheckResourceAttr("data.seca_public_ip.test", "state", "active"),
				),
			},
		},
	})
}
