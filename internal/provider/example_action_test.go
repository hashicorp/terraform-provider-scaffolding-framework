// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccExampleAction(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExampleActionConfig,
				PostApplyFunc: func() {
					// Test the results of an action operation.
					// Actions should not affect existing resources managed
					// by Terraform, so testing should be scoped to real-world side effects,
					// rather than Terraform plan or state values.
					//
					// For example, an action that changes the contents of a local file
					// could test the contents of that file:

					// resultContent, err := os.ReadFile(f)
					// if err != nil {
					//	 t.Errorf("Error occurred while reading file at path: %s\n, error: %s\n", f, err)
					// }
					//
					// if string(resultContent) != updatedContent {
					//	 t.Errorf("Expected file content %s\n, got: %s\n", updatedContent, resultContent)
					// }
				},
			},
		},
	})
}

const testAccExampleActionConfig = `
resource "terraform_data" "test" {
	input = "fake-string"

	lifecycle {
		action_trigger {
		  events  = [before_create] # action triggers before resource creation
		  actions = [action.scaffolding_example.test]
		}
	}
}

action "scaffolding_example" "test" {
	config {
		configurable_attribute = "example"
	}
}`
