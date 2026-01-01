# Known Issues

## Spurious In-Place Updates

### Private Resources (`sse_private_resource`)

Users may observe `tofu plan` or `terraform plan` showing in-place updates for `sse_private_resource` even when no configuration changes have been made. This typically manifests as the addition of computed fields (like `external_fqdn_prefix`) inside `access_types` that were not explicitly defined in the configuration but are returned by the API.

**Example:**
```hcl
  # sse_private_resource.ERP will be updated in-place
  ~ resource "sse_private_resource" "ERP" {
        id   = "284044"
        name = "ERP"

      ~ access_types {
          + external_fqdn_prefix     = "erp-8337932"
            # (4 unchanged attributes hidden)
        }
    }
```

### Access Rules (`sse_access_rule`)

Similar behavior may be observed with `sse_access_rule` resources. The provider may detect differences between the local state and the API response for certain fields, causing Terraform/OpenTofu to propose an in-place update to reconcile the state, even if the logical configuration remains unchanged.
