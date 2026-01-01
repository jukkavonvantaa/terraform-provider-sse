// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPrivateResourceGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPrivateResourceGroupResourceConfig("test-prg"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sse_private_resource_group.test", "name", "test-prg"),
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
				Config: testAccPrivateResourceGroupResourceConfig("test-prg-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sse_private_resource_group.test", "name", "test-prg-updated"),
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
