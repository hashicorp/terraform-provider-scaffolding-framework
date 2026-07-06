package acctest

import (
	"context"
	"fmt"
	"testing"

	"github.com/eu-sovereign-cloud/go-sdk/secapi"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckInternetGatewayDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := testAccRegionalClient(ctx)
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "seca_internet_gateway" {
			continue
		}

		wref := secapi.WorkspaceReference{
			Tenant:    secapi.TenantID(testAccTenant),
			Workspace: secapi.WorkspaceID(rs.Primary.Attributes["workspace_id"]),
			Name:      rs.Primary.Attributes["name"],
		}

		_, err := client.NetworkV1.GetInternetGateway(ctx, wref)
		if err == secapi.ErrResourceNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking internet gateway %q was destroyed: %w", wref.Name, err)
		}
		return fmt.Errorf("internet gateway %q still exists after destroy", wref.Name)
	}

	return nil
}

func testAccInternetGatewayResourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_workspace" "test" {
  name = "workspace-1"
}
resource "seca_internet_gateway" "test" {
  name         = "internet-gateway-1"
  workspace_id = seca_workspace.test.name

  labels = %s
}
`, formatLabels(labels))
}

func testAccInternetGatewayDataSourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_workspace" "test" {
  name = "workspace-1"
}
resource "seca_internet_gateway" "test" {
  name         = "internet-gateway-1"
  workspace_id = seca_workspace.test.name

  labels = %s
}
data "seca_internet_gateway" "test" {
  name         = "internet-gateway-1"
  workspace_id = seca_workspace.test.name
}`, formatLabels(labels))
}

func TestAccInternetGateway(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInternetGatewayResourceConfig(map[string]string{"env": "dev"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_internet_gateway.test", "name", "internet-gateway-1"),
					resource.TestCheckResourceAttr("seca_internet_gateway.test", "workspace_id", "workspace-1"),
					resource.TestCheckResourceAttr("seca_internet_gateway.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("seca_internet_gateway.test", "region", testAccRegion),
					resource.TestCheckResourceAttr("seca_internet_gateway.test", "labels.env", "dev"),
				),
			},
			{
				Config: testAccInternetGatewayResourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_internet_gateway.test", "name", "internet-gateway-1"),
					resource.TestCheckResourceAttr("seca_internet_gateway.test", "labels.env", "prod"),
				),
			},
			{
				ResourceName:      "seca_internet_gateway.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "workspace-1/internet-gateway-1",
			},
			{
				Config: testAccInternetGatewayDataSourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_internet_gateway.test", "name", "internet-gateway-1"),
					resource.TestCheckResourceAttr("seca_internet_gateway.test", "workspace_id", "workspace-1"),

					resource.TestCheckResourceAttr("data.seca_internet_gateway.test", "name", "internet-gateway-1"),
					resource.TestCheckResourceAttr("data.seca_internet_gateway.test", "workspace_id", "workspace-1"),
					resource.TestCheckResourceAttr("data.seca_internet_gateway.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("data.seca_internet_gateway.test", "region", testAccRegion),
					resource.TestCheckResourceAttr("data.seca_internet_gateway.test", "state", "active"),
				),
			},
		},
	})
}
