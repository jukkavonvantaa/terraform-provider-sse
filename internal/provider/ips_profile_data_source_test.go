// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIPSProfileDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "sse_ips_profiles" "all" {}

data "sse_ips_profile" "test" {
  name = data.sse_ips_profiles.all.ips_profiles[0].name
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify plural data source
					resource.TestCheckResourceAttrSet("data.sse_ips_profiles.all", "ips_profiles.#"),
					
					// Verify singular data source
					resource.TestCheckResourceAttrSet("data.sse_ips_profile.test", "id"),
					resource.TestCheckResourceAttrSet("data.sse_ips_profile.test", "name"),
					resource.TestCheckResourceAttrSet("data.sse_ips_profile.test", "is_default"),
					resource.TestCheckResourceAttrSet("data.sse_ips_profile.test", "system_mode"),
				),
			},
		},
	})
}
