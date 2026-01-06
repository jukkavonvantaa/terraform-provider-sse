
data "sse_tenant_controls_profile" "example" {
  name = "Test Tenant Profile"
}

data "sse_identity" "Engineering" {
  name = "Engineering (example.net\\Engineering)"
}
data "sse_identity" "Inventory" {
  name = "Inventory (example.net\\Inventory)"
}

data "sse_ips_profile" "Tofudemo" {
  name = "Tofudemo"
}

resource "sse_service_object" "service_object-1" {
  name        = "example-Terraform Service Object-1"
  description = "Managed by Terraform"
  protocol    = "tcp"
  ports       = ["80", "441", "8080-8090"]
}

resource "sse_private_resource" "example-rdp1" {
  name = "example-rdp-1"

  access_types {
    external_fqdn_prefix     = "example-rdp-1-8337022"
    protocol                 = "rdp-tcp"
    ssl_verification_enabled = true
    type                     = "browser"
  }
  access_types {
    reachable_addresses = [
      "10.10.22.134",
    ]
    ssl_verification_enabled = false
    type                     = "client"
  }

  resource_addresses {
    destination_addr = [
      "10.10.22.1",
    ]
    protocol_ports {
      ports    = "3389"
      protocol = "rdp-tcp"
    }
  }
}

# Private access rule
resource "sse_access_rule" "example-rdp1" {
  action     = "allow"
  is_enabled = true
  name       = "example-rdp-1"

  # identity
  rule_conditions {
    attribute_name     = "umbrella.source.identity_ids"
    attribute_operator = "INTERSECT"
    attribute_value = jsonencode(
      [
        data.sse_identity.Engineering.id,
        data.sse_identity.Inventory.id,
      ]
    )
  }

  # private resource
  rule_conditions {
    attribute_name     = "umbrella.destination.private_resource_ids"
    attribute_operator = "IN"
    attribute_value = jsonencode(
      [
        sse_private_resource.example-rdp1.resource_id
      ]
    )
  }

  rule_settings {
    setting_name  = "umbrella.logLevel"
    setting_value = "LOG_ALL"
  }
  rule_settings {
    setting_name  = "umbrella.default.traffic"
    setting_value = "PRIVATE_NETWORK"
  }
}

# Destination list
resource "sse_destination_list" "example-list-1" {
  name           = "example-list-1"
  access         = "block"
  is_global      = false
  bundle_type_id = 2
  destinations = [
    {
      destination = "example.com"
      type        = "DOMAIN"
    }
  ]
}

# Network object
resource "sse_network_object" "example-obj-1" {
  name        = "example-obj-1"
  description = "Managed by Terraform"
  type        = "network"
  addresses   = ["192.168.1.0/24"]
}

# Internet access rule
resource "sse_access_rule" "example-inet-1" {
  name        = "example-inet-1"
  description = "Uses Network Objects, Destination Lists, and Private Resources"
  action      = "allow"
  is_enabled  = true

  # Source: Network Object
  rule_conditions {
    attribute_name     = "umbrella.source.networkObjectIds"
    attribute_operator = "IN"
    attribute_value = jsonencode([
      # Use object_id (integer) instead of id (string)
      sse_network_object.example-obj-1.object_id
    ])
  }

  # Destination: Destination List
  rule_conditions {
    attribute_name     = "umbrella.destination.destination_list_ids"
    attribute_operator = "INTERSECT"
    attribute_value = jsonencode([
      # Use list_id (integer) instead of id (string)
      sse_destination_list.example-list-1.list_id
    ])
  }

  # Destination: Service object
  rule_conditions {
    attribute_name     = "umbrella.destination.serviceObjectIds"
    attribute_operator = "IN"
    attribute_value = jsonencode([
      # Use object_id (integer) instead of id (string)
      sse_service_object.service_object-1.object_id
    ])
  }

  # Tenant control
  rule_settings {
    setting_name  = "sse.tenantControlProfileId"
    setting_value = data.sse_tenant_controls_profile.example.id
  }

  # IPS profile
  rule_settings {
    setting_name  = "umbrella.posture.ipsProfileId"
    setting_value = data.sse_ips_profile.Tofudemo.id
  }

  rule_settings {
    setting_name  = "umbrella.logLevel"
    setting_value = "LOG_ALL"
  }

  rule_settings {
    setting_name  = "umbrella.default.traffic"
    setting_value = "PUBLIC_INTERNET"
  }
}
