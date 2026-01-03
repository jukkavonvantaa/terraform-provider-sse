resource "sse_access_rule" "example" {
  name        = "Terraform Example Rule"
  description = "Created via Terraform"
  action      = "allow"
  is_enabled = true

  rule_conditions {
    attribute_name     = "umbrella.destination.all"
    attribute_value    = "true"
    attribute_operator = "="
  }

  rule_conditions {
    attribute_name     = "umbrella.source.all"
    attribute_value    = "true"
    attribute_operator = "="
  }

  rule_settings {
    setting_name  = "umbrella.default.traffic"
    setting_value = "PUBLIC_INTERNET"
  }
}

resource "sse_access_rule" "complex_example" {
  name        = "Terraform Complex Rule"
  description = "Complex rule with IP/Port/Protocol"
  action      = "allow"
  is_enabled  = false

  rule_conditions {
    attribute_name = "umbrella.destination.composite_inline_ip"
    attribute_value = jsonencode([
      {
        "ip" : ["1.2.3.0/24"],
        "port" : ["80"],
        "protocol" : "TCP"
      }
    ])
    attribute_operator = "IN"
  }

  rule_conditions {
    attribute_name     = "umbrella.source.all"
    attribute_value    = "true"
    attribute_operator = "="
  }

  rule_settings {
    setting_name  = "umbrella.default.traffic"
    setting_value = "PUBLIC_INTERNET"
  }
}

# Example using identities, applications and category lists
data "sse_identities" "all" {}
data "sse_application" "facebook" {
  name = "Facebook"
}
data "sse_content_category_lists" "all" {}
data "sse_security_profile" "web_profile" {
  name = "Web Profile"
}
data "sse_ips_profile" "standard" {
  name = "Standard IPS Profile"
}

resource "sse_access_rule" "identity_example" {
  name        = "Rule with Dynamic Identities"
  description = "Uses identities data source"
  action      = "allow"
  is_enabled  = true

  rule_conditions {
    attribute_name     = "umbrella.source.identity_ids"
    attribute_operator = "INTERSECT"
    attribute_value    = jsonencode(
      [
        # Use the data source to find IDs dynamically
        [for i in data.sse_identities.all.identities : i.id if i.label == "Larry Laffer (llaffer@example.net)"][0],
        [for i in data.sse_identities.all.identities : i.id if i.label == "Tove Jansson (tjansson@example.net)"][0]
      ]
    )
  }

  rule_conditions {
    attribute_name     = "umbrella.destination.application_ids"
    attribute_operator = "INTERSECT"
    attribute_value    = jsonencode([
      data.sse_application.facebook.id
    ])
  }

  rule_conditions {
    attribute_name     = "umbrella.destination.category_list_ids"
    attribute_operator = "INTERSECT"
    attribute_value    = jsonencode([
      [for list in data.sse_content_category_lists.all.content_category_lists : list.id if list.name == "Banned content"][0]
    ])
  }

  rule_settings {
    setting_name  = "umbrella.default.traffic"
    setting_value = "PUBLIC_INTERNET"
  }

  rule_settings {
    setting_name  = "umbrella.posture.webProfileId"
    setting_value = data.sse_security_profile.web_profile.id
  }

  rule_settings {
    setting_name  = "umbrella.posture.ipsProfileId"
    setting_value = data.sse_ips_profile.standard.id
  }
}
