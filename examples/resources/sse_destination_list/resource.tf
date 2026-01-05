resource "sse_destination_list" "example" {
  name      = "Terraform Destination List"
  access    = "allow"
  is_global = false
  # bundle_type_id: 1 = DNS (Domains only), 2 = Web (Domains, URLs, IPs), 4 = SAML Bypass
  bundle_type_id = 1
  destinations = [
    {
      destination = "example.com"
      type        = "domain"
      comment     = "Example Domain"
    },
    {
      destination = "1.1.1.1"
      type        = "ipv4"
      comment     = "Example IP"
    }
  ]
}
