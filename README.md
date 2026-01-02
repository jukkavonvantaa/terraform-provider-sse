# Terraform/OpenTofu Provider for Cisco Secure Access (SSE)

This is a Terraform/OpenTofu provider for **Cisco Secure Access (SSE)**.

It allows you to manage resources such as:
- Network Objects
- Destination Lists
- Access Policy Rules
- Private Resources & Groups
- Service Objects
- Network Tunnel Groups (Data Source)
- Identities (Data Source)
- Resource Connector Groups (Resource & Data Source)
- Applications (Data Source)
- Application Categories (Data Source)
- Content Category Lists (Data Source)

> **Note:** As of January 1, 2026, there is no official Terraform provider for Cisco Secure Access. This project was "vibe coded" using **Gemini 3 Pro** to fill that gap.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0 or [OpenTofu](https://opentofu.org/)
- [Go](https://golang.org/doc/install) >= 1.24

## Configuration

The provider requires API credentials to communicate with Cisco Secure Access. You can set these via environment variables:

```bash
export SSE_CLIENT_KEY="your_client_key"
export SSE_CLIENT_SECRET="your_client_secret"
# Optional: Defaults to "us". Use "eu" for Europe region.
# export SSE_REGION="us"
# Optional: Defaults to https://api.sse.cisco.com/auth/v2/token
# export SSE_TOKEN_URL="https://api.sse.cisco.com/auth/v2/token"
```

## API Rate Limiting and Locking

The Cisco Secure Access API enforces strict rate limits and locks the ruleset during modifications.
To handle this, the provider implements automatic retries with exponential backoff for:
- **429 Too Many Requests**: Retries after a delay.
- **409 Conflict (Ruleset Locked)**: Retries if the API reports the ruleset is locked by another process.

If you still encounter issues with large configurations, consider reducing the parallelism of Terraform/OpenTofu:

```bash
tofu apply -parallelism=1
```

## Installation

To install the provider locally for development:

```bash
cd terraform-provider-sse
go install .
```

This will build the binary and install it to your `$GOPATH/bin`.

## Developer Mode

To use the locally built provider without publishing it to a registry, you can configure Terraform/OpenTofu to use the binary from your Go bin directory.

Create or edit your `$HOME/.terraformrc` (or `%APPDATA%\terraform.rc` on Windows) file:

```hcl
provider_installation {

  dev_overrides {
      "registry.opentofu.org/hashicorp/sse" = "/Users/<username>/go/bin"
      "registry.terraform.io/cisco/sse" = "/Users/<username>/go/bin"
  }

  direct {}
}
```

Replace `<username>` with your actual username.

Note that in developer mode you don't use `tofu init` command.

## Usage Example

```hcl
provider "sse" {}

resource "sse_network_object" "example" {
  name      = "example-network"
  type      = "host"
  addresses = ["192.168.1.1"]
}

resource "sse_destination_list" "blocklist" {
  name   = "My Block List"
  access = "block"
  destinations = [
    {
      destination = "badsite.com"
      type        = "domain"
    }
  ]
}

resource "sse_connector_group" "nyc_office" {
  name        = "NYC Office Connector Group"
  location    = "us-east-1"
  environment = "aws"
}

# Fetch all connector groups
data "sse_connector_groups" "all" {}
```

## License

This project is licensed under the **European Union Public Licence, version 1.2 (EUPL-1.2)**.
See the [LICENSE](LICENSE) file for details or visit https://opensource.org/license/eupl-1-2.

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
$ make testacc
```

## OpenAPI Specifications

The OpenAPI specifications for the Cisco Secure Access API have been included in this repository to assist with development and AI-assisted coding. These files were downloaded from [developer.cisco.com](https://developer.cisco.com) in December 2025.

---
May the power of AI save you from the perils of ClickOps.

Best Regards,
Gemini 3 Pro

