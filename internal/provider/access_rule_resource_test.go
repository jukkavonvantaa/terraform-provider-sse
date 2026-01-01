// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAccessRuleResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAccessRuleResourceConfig(fmt.Sprintf("test-access-rule-%s", rName), "allow", "1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sse_access_rule.test", "name", fmt.Sprintf("test-access-rule-%s", rName)),
					resource.TestCheckResourceAttr("sse_access_rule.test", "action", "allow"),
					resource.TestCheckResourceAttr("sse_access_rule.test", "priority", "1"),
					resource.TestCheckResourceAttr("sse_access_rule.test", "is_enabled", "false"),
					resource.TestCheckResourceAttrSet("sse_access_rule.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sse_access_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore fields that might be returned differently or are not in the state exactly as expected
				// For example, rule_conditions and rule_settings might be reordered or formatted differently.
				// For now, let's try verifying everything and see if it fails.
				// If it fails on complex objects, we might need to ignore them.
				// ImportStateVerifyIgnore: []string{"rule_conditions", "rule_settings"},
			},
			// Import by Name testing
			{
				ResourceName:      "sse_access_rule.test",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("test-access-rule-%s", rName),
				ImportStateVerify: true,
				// ImportStateVerifyIgnore: []string{"rule_conditions", "rule_settings"},
			},
			// Update and Read testing
			{
				Config: testAccAccessRuleResourceConfig(fmt.Sprintf("test-access-rule-updated-%s", rName), "block", "2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sse_access_rule.test", "name", fmt.Sprintf("test-access-rule-updated-%s", rName)),
					resource.TestCheckResourceAttr("sse_access_rule.test", "action", "block"),
					resource.TestCheckResourceAttr("sse_access_rule.test", "priority", "2"),
				),
			},
		},
	})
}

func testAccAccessRuleResourceConfig(name, action, priority string) string {
	return fmt.Sprintf(`
resource "sse_access_rule" "test" {
  name        = %[1]q
  action      = %[2]q
  priority    = %[3]s
  is_enabled  = false
  description = "Test Access Rule"

  rule_conditions {
    attribute_name     = "umbrella.source.all"
    attribute_value    = "true"
    attribute_operator = "="
  }

  rule_conditions {
    attribute_name     = "umbrella.destination.composite_inline_ip"
    attribute_value    = "[{\"ip\":[\"1.2.3.0/24\"],\"port\":[\"80\"],\"protocol\":\"TCP\"}]"
    attribute_operator = "IN"
  }

  rule_settings {
    setting_name  = "umbrella.default.traffic"
    setting_value = "PUBLIC_INTERNET"
  }
}
`, name, action, priority)
}
