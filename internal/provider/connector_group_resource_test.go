package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccConnectorGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccConnectorGroupResourceConfig("test-connector-group", "us-west-2", "aws"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sse_connector_group.test", "name", "test-connector-group"),
					resource.TestCheckResourceAttr("sse_connector_group.test", "location", "us-west-2"),
					resource.TestCheckResourceAttr("sse_connector_group.test", "environment", "aws"),
					resource.TestCheckResourceAttrSet("sse_connector_group.test", "id"),
					resource.TestCheckResourceAttrSet("sse_connector_group.test", "provisioning_key"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sse_connector_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// ImportState testing by Name
			{
				ResourceName:      "sse_connector_group.test",
				ImportState:       true,
				ImportStateId:     "test-connector-group",
				ImportStateVerify: true,
			},
			// Update testing
			{
				Config: testAccConnectorGroupResourceConfig("test-connector-group-updated", "us-east-1", "aws"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sse_connector_group.test", "name", "test-connector-group-updated"),
					resource.TestCheckResourceAttr("sse_connector_group.test", "location", "us-east-1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccConnectorGroupResourceConfig(name, location, environment string) string {
	return fmt.Sprintf(`
resource "sse_connector_group" "test" {
  name        = %[1]q
  location    = %[2]q
  environment = %[3]q
}
`, name, location, environment)
}
