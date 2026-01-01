// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetworkObjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNetworkObjectResourceConfig("test-net-obj", "1.1.1.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sse_network_object.test", "name", "test-net-obj"),
					resource.TestCheckResourceAttr("sse_network_object.test", "type", "host"),
					resource.TestCheckResourceAttr("sse_network_object.test", "addresses.0", "1.1.1.1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sse_network_object.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccNetworkObjectResourceConfig("test-net-obj-updated", "2.2.2.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sse_network_object.test", "name", "test-net-obj-updated"),
					resource.TestCheckResourceAttr("sse_network_object.test", "addresses.0", "2.2.2.2"),
				),
			},
		},
	})
}

func testAccNetworkObjectResourceConfig(name, address string) string {
	return fmt.Sprintf(`
resource "sse_network_object" "test" {
  name      = %[1]q
  type      = "host"
  addresses = [%[2]q]
}
`, name, address)
}
