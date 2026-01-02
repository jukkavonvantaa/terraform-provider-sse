# Fetch all connector groups
data "sse_connector_groups" "all" {}

# Output the list of groups
output "all_connector_groups" {
  value = data.sse_connector_groups.all.connector_groups
}
