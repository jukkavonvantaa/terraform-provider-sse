# Find a specific identity by name (label)
data "sse_identity" "user" {
  name = "Larry Laffer (llaffer@example.net)"
}

output "user_id" {
  value = data.sse_identity.user.id
}

output "user_details" {
  value = data.sse_identity.user
}
