// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDestinationListResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDestinationListResourceConfig(fmt.Sprintf("test-dest-list-%s", rName), "allow"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sse_destination_list.test", "name", fmt.Sprintf("test-dest-list-%s", rName)),
					resource.TestCheckResourceAttr("sse_destination_list.test", "access", "allow"),
					resource.TestCheckResourceAttr("sse_destination_list.test", "is_global", "false"),
					resource.TestCheckResourceAttr("sse_destination_list.test", "bundle_type_id", "2"),
					resource.TestCheckResourceAttrSet("sse_destination_list.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sse_destination_list.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The ID is computed, so we can't verify it in the import state check easily if we don't know it.
				// But ImportStateVerify=true checks that the state in Terraform matches the state from the API.
				// We might need to ignore some fields if they are not returned by the API or formatted differently.
				ImportStateVerifyIgnore: []string{"modified_at", "created_at"},
			},
			// Update and Read testing
			{
				Config: testAccDestinationListResourceConfig(fmt.Sprintf("test-dest-list-updated-%s", rName), "block"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sse_destination_list.test", "name", fmt.Sprintf("test-dest-list-updated-%s", rName)),
					resource.TestCheckResourceAttr("sse_destination_list.test", "access", "block"),
				),
			},
		},
	})
}

func testAccDestinationListResourceConfig(name, access string) string {
	return fmt.Sprintf(`
resource "sse_destination_list" "test" {
  name           = %[1]q
  access         = %[2]q
  is_global      = false
  bundle_type_id = 2
}
`, name, access)
}
