---
page_title: "sse_connector_group Resource - sse"
subcategory: ""
description: |-
  Manages a Resource Connector Group.
---

# sse_connector_group (Resource)

Manages a Resource Connector Group.

## Example Usage

```terraform
resource "sse_connector_group" "nyc_office" {
  name        = "NYC Office Connector Group"
  location    = "us-east-1"
  environment = "aws"
}
```

## Schema

### Required

- `name` (String) The name of the Connector Group.
- `location` (String) The region where the Resource Connector Group is available (e.g., us-west-2).
- `environment` (String) The type of cloud-native runtime environment (e.g., aws, azure, container, esx). **Note:** This field is immutable and requires replacement if changed.

### Read-Only

- `id` (Number) The ID of the Connector Group.
- `provisioning_key` (String, Sensitive) The provisioning key for the Connector Group.
- `provisioning_key_expires_at` (String) The expiration time of the provisioning key.
- `base_image_download_url` (String) The URL to download the base image.
- `status` (String) The status of the Connector Group.

## Import

Import is supported using the following syntax:

```shell
# Import by ID
terraform import sse_connector_group.example 123456

# Import by Name
terraform import sse_connector_group.example "NYC Office Connector Group"
```
