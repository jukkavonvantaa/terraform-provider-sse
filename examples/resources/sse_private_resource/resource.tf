resource "sse_private_resource_group" "example" {
  name        = "Example Group for Resource"
  description = "Group for the private resource example"
}

resource "sse_private_resource" "example" {
  name        = "Example Private Resource"
  description = "Managed by Terraform"

  resource_group_ids = [sse_private_resource_group.example.id]

  access_types {
    type                     = "network"
    protocol                 = "TCP"
    ssl_verification_enabled = false
  }

  resource_addresses {
    destination_addr = ["192.168.1.100"]
    protocol_ports {
      protocol = "TCP"
      ports    = "8080"
    }
  }
}
