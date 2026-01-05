data "sse_tenant_controls_profile" "example" {
  name = "Global Tenant Controls"
}

resource "sse_access_rule" "tenant_control_rule" {
  name        = "Rule with Tenant Control"
  description = "Access rule using Tenant Control Profile"
  action      = "allow"
  priority    = 1
  is_enabled  = true

  rule_conditions {
    attribute_name     = "umbrella.destination.application_ids"
    attribute_value    = jsonencode([710]) # Example Application ID
    attribute_operator = "INTERSECT"
  }

  rule_settings {
    setting_name  = "umbrella.default.traffic"
    setting_value = "PUBLIC_INTERNET"
  }

  rule_settings {
    setting_name  = "sse.tenantControlProfileId"
    setting_value = data.sse_tenant_controls_profile.example.id
  }
}
