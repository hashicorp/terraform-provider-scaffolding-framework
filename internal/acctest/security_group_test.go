package acctest

import (
	"context"
	"fmt"
	"testing"

	"github.com/eu-sovereign-cloud/go-sdk/secapi"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckSecurityGroupDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := testAccRegionalClient(ctx)
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "seca_security_group" {
			continue
		}

		wref := secapi.WorkspaceReference{
			Tenant:    secapi.TenantID(testAccTenant),
			Workspace: secapi.WorkspaceID(rs.Primary.Attributes["workspace_id"]),
			Name:      rs.Primary.Attributes["name"],
		}

		_, err := client.NetworkV1.GetSecurityGroup(ctx, wref)
		if err == secapi.ErrResourceNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking security group %q was destroyed: %w", wref.Name, err)
		}
		return fmt.Errorf("security group %q still exists after destroy", wref.Name)
	}

	return nil
}

func testAccSecurityGroupResourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_workspace" "test" {
  name = "workspace-1"
}
resource "seca_security_group" "test" {
  name         = "security-group-1"
  workspace_id = seca_workspace.test.name

  rules = [
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        list = [80, 443]
      }
      source_refs = []
    }
  ]
  labels = %s
}
`, formatLabels(labels))
}

func testAccSecurityGroupUpdateConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_workspace" "test" {
  name = "workspace-1"
}
resource "seca_security_group" "test" {
  name         = "security-group-1"
  workspace_id = seca_workspace.test.name

  rules = [
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        list = [80, 443]
      }
      source_refs = []
    },
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        from = 22
      }
      source_refs = ["55.44.33.11"]
    }
  ]
  labels = %s
}
`, formatLabels(labels))
}

func testAccSecurityGroupDataSourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_workspace" "test" {
  name = "workspace-1"
}
resource "seca_security_group" "test" {
  name         = "security-group-1"
  workspace_id = seca_workspace.test.name

  rules = [
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        list = [80, 443]
      }
      source_refs = []
    }
  ]
  labels = %s
}
data "seca_security_group" "test" {
  name         = "security-group-1"
  workspace_id = seca_workspace.test.name
}`, formatLabels(labels))
}

func TestAccSecurityGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupResourceConfig(map[string]string{"env": "dev"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_security_group.test", "name", "security-group-1"),
					resource.TestCheckResourceAttr("seca_security_group.test", "workspace_id", "workspace-1"),
					resource.TestCheckResourceAttr("seca_security_group.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("seca_security_group.test", "region", testAccRegion),
					resource.TestCheckResourceAttr("seca_security_group.test", "rules.#", "1"),
					resource.TestCheckResourceAttr("seca_security_group.test", "rules.0.direction", "ingress"),
					resource.TestCheckResourceAttr("seca_security_group.test", "rules.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("seca_security_group.test", "labels.env", "dev"),
				),
			},
			{
				Config: testAccSecurityGroupUpdateConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_security_group.test", "name", "security-group-1"),
					resource.TestCheckResourceAttr("seca_security_group.test", "rules.#", "2"),
					resource.TestCheckResourceAttr("seca_security_group.test", "labels.env", "prod"),
				),
			},
			{
				ResourceName:      "seca_security_group.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "workspace-1/security-group-1",
			},
			{
				Config: testAccSecurityGroupDataSourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_security_group.test", "name", "security-group-1"),
					resource.TestCheckResourceAttr("seca_security_group.test", "workspace_id", "workspace-1"),

					resource.TestCheckResourceAttr("data.seca_security_group.test", "name", "security-group-1"),
					resource.TestCheckResourceAttr("data.seca_security_group.test", "workspace_id", "workspace-1"),
					resource.TestCheckResourceAttr("data.seca_security_group.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("data.seca_security_group.test", "region", testAccRegion),
					resource.TestCheckResourceAttr("data.seca_security_group.test", "state", "active"),
				),
			},
		},
	})
}
