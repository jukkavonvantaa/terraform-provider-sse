# Find a specific application by name
data "sse_application" "facebook" {
  name = "Facebook"
}

output "facebook_details" {
  value = data.sse_application.facebook
}

output "facebook_id" {
  value = data.sse_application.facebook.id
}
