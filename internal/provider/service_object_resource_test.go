// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServiceObjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccServiceObjectResourceConfig("test-svc-obj", "TCP", "80", "443"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sse_service_object.test", "name", "test-svc-obj"),
					resource.TestCheckResourceAttr("sse_service_object.test", "protocol", "TCP"),
					resource.TestCheckResourceAttr("sse_service_object.test", "ports.0", "80"),
					resource.TestCheckResourceAttr("sse_service_object.test", "ports.1", "443"),
					resource.TestCheckResourceAttrSet("sse_service_object.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sse_service_object.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccServiceObjectResourceConfig("test-svc-obj-updated", "UDP", "53"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sse_service_object.test", "name", "test-svc-obj-updated"),
					resource.TestCheckResourceAttr("sse_service_object.test", "protocol", "UDP"),
					resource.TestCheckResourceAttr("sse_service_object.test", "ports.0", "53"),
				),
			},
		},
	})
}

func testAccServiceObjectResourceConfig(name, protocol string, ports ...string) string {
	portsConfig := ""
	for _, p := range ports {
		portsConfig += fmt.Sprintf(`"%s",`, p)
	}
	if len(portsConfig) > 0 {
		portsConfig = portsConfig[:len(portsConfig)-1] // Remove trailing comma
	}

	return fmt.Sprintf(`
resource "sse_service_object" "test" {
  name        = "%s"
  description = "Acceptance Test Service Object"
  protocol    = "%s"
  ports       = [%s]
}
`, name, protocol, portsConfig)
}
