package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIdentitiesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "sse_identities" "all" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify that we get back a list of identities
					resource.TestCheckResourceAttrSet("data.sse_identities.all", "identities.#"),
				),
			},
		},
	})
}
