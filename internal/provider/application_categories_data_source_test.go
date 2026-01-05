// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApplicationCategoriesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "sse_application_categories" "all" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify that we get back a list of application categories
					resource.TestCheckResourceAttrSet("data.sse_application_categories.all", "application_categories.#"),
				),
			},
		},
	})
}
