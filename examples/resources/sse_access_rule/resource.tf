resource "sse_access_rule" "example" {
  name        = "Terraform Example Rule"
  description = "Created via Terraform"
  action      = "allow"
  # Priority must be sequential and within the range of existing rules.
  # If you have 5 rules, the next priority can be 6.
  priority   = 1
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
  priority    = 2
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

# Example using dynamic identities
data "sse_identities" "all" {}

resource "sse_access_rule" "identity_example" {
  name        = "Rule with Dynamic Identities"
  description = "Uses identities data source"
  action      = "allow"
  priority    = 3
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
    attribute_name     = "umbrella.destination.all"
    attribute_value    = "true"
    attribute_operator = "="
  }

  rule_settings {
    setting_name  = "umbrella.default.traffic"
    setting_value = "PUBLIC_INTERNET"
  }
}
