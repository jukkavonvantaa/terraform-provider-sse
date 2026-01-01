resource "sse_service_object" "example" {
  name        = "Terraform Service Object"
  description = "Managed by Terraform"
  protocol    = "tcp"
  ports       = ["80", "443", "8080-8090"]
}
