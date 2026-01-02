package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccContentCategoryListsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "sse_content_category_lists" "all" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify that we get back a list of content category lists
					resource.TestCheckResourceAttrSet("data.sse_content_category_lists.all", "content_category_lists.#"),
				),
			},
		},
	})
}
