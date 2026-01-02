terraform {
  required_providers {
    sse = {
      source = "registry.terraform.io/cisco/sse"
    }
  }
}

provider "sse" {
  # client_id     = "..." # Set via SSE_CLIENT_ID env var
  # client_secret = "..." # Set via SSE_CLIENT_SECRET env var
  # region        = "..." # Set via SSE_REGION env var (optional, defaults to "us")
}

resource "sse_network_object" "example" {
  name        = "terraform-example-object"
  description = "Managed by Terraform"
  type        = "host"
  addresses   = ["192.168.1.100"]
}
