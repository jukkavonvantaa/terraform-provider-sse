# Fetch all identities
data "sse_identities" "all" {}

# Output all identities
output "all_identities" {
  value = data.sse_identities.all.identities
}

# Find a specific identity ID by label
output "larry_laffner_id" {
  value = [for i in data.sse_identities.all.identities : i.id if i.label == "Larry Laffner (llaffner@example.net)"][0]
}
