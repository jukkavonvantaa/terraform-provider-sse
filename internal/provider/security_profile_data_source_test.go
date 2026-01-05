// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSecurityProfileDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "sse_security_profiles" "all" {}

data "sse_security_profile" "test" {
  name = data.sse_security_profiles.all.security_profiles[0].name
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify plural data source
					resource.TestCheckResourceAttrSet("data.sse_security_profiles.all", "security_profiles.#"),
					
					// Verify singular data source
					resource.TestCheckResourceAttrSet("data.sse_security_profile.test", "id"),
					resource.TestCheckResourceAttrSet("data.sse_security_profile.test", "name"),
					resource.TestCheckResourceAttrSet("data.sse_security_profile.test", "is_default"),
					resource.TestCheckResourceAttrSet("data.sse_security_profile.test", "priority"),
				),
			},
		},
	})
}
