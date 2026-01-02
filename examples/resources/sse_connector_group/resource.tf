# Create a new Connector Group
resource "sse_connector_group" "nyc_office" {
  name        = "NYC Office Connector Group"
  location    = "us-east-1"
  environment = "aws"
}

# Output the provisioning key (Sensitive)
output "provisioning_key" {
  value     = sse_connector_group.nyc_office.provisioning_key
  sensitive = true
}
