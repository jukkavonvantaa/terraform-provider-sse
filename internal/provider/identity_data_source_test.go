// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIdentityDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "sse_identities" "all" {}

data "sse_identity" "test" {
  name = data.sse_identities.all.identities[0].label
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sse_identity.test", "id"),
					resource.TestCheckResourceAttrSet("data.sse_identity.test", "label"),
					resource.TestCheckResourceAttrSet("data.sse_identity.test", "type"),
				),
			},
		},
	})
}
