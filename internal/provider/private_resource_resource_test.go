// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPrivateResourceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPrivateResourceResourceConfig("test-pr"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sse_private_resource.test", "name", "test-pr"),
					resource.TestCheckResourceAttrSet("sse_private_resource.test", "id"),
					resource.TestCheckResourceAttr("sse_private_resource.test", "resource_addresses.0.destination_addr.0", "1.2.3.4"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sse_private_resource.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccPrivateResourceResourceConfig("test-pr-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sse_private_resource.test", "name", "test-pr-updated"),
				),
			},
		},
	})
}

func testAccPrivateResourceResourceConfig(name string) string {
	return `
resource "sse_private_resource" "test" {
  name        = "` + name + `"
  description = "Acceptance Test Private Resource"
  resource_addresses {
    destination_addr = ["1.2.3.4"]
    protocol_ports {
      protocol = "tcp"
      ports    = "80"
    }
  }
  access_types {
    type = "network"
    ssl_verification_enabled = false
  }
}
`
}
