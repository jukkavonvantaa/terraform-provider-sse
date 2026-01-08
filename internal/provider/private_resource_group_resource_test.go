// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPrivateResourceGroupResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	name := fmt.Sprintf("test-prg-%s", rName)
	updatedName := fmt.Sprintf("test-prg-updated-%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPrivateResourceGroupResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sse_private_resource_group.test", "name", name),
					resource.TestCheckResourceAttrSet("sse_private_resource_group.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sse_private_resource_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccPrivateResourceGroupResourceConfig(updatedName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sse_private_resource_group.test", "name", updatedName),
				),
			},
		},
	})
}

func testAccPrivateResourceGroupResourceConfig(name string) string {
	return `
resource "sse_private_resource_group" "test" {
  name        = "` + name + `"
  description = "Acceptance Test Private Resource Group"
}
`
}
