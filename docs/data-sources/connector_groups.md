---
page_title: "sse_connector_groups Data Source - sse"
subcategory: ""
description: |-
  Fetches the list of Resource Connector Groups.
---

# sse_connector_groups (Data Source)

Fetches the list of Resource Connector Groups.

## Example Usage

```terraform
data "sse_connector_groups" "all" {}

output "all_connector_groups" {
  value = data.sse_connector_groups.all.connector_groups
}
```

## Schema

### Read-Only

- `connector_groups` (Attributes List) (see below for nested schema)

<a id="nestedatt--connector_groups"></a>
### Nested Schema for `connector_groups`

Read-Only:

- `id` (Number) The ID of the Connector Group.
- `name` (String) The name of the Connector Group.
- `location` (String) The region where the Resource Connector Group is available.
- `environment` (String) The type of cloud-native runtime environment.
- `status` (String) The status of the Connector Group.
- `connectors_count` (Number) Total number of connectors.
- `connected_connectors_count` (Number) Number of connected connectors.
- `disconnected_connectors_count` (Number) Number of disconnected connectors.
