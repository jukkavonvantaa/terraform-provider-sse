resource "sse_network_object" "example" {
  name        = "Terraform Network Object"
  description = "Managed by Terraform"
  type        = "network"
  addresses   = ["192.168.1.0/24", "10.0.0.0/8"]
}
