# Find a specific security profile by name
data "sse_security_profile" "web_profile" {
  name = "Web Profile"
}

output "profile_id" {
  value = data.sse_security_profile.web_profile.id
}

output "profile_details" {
  value = data.sse_security_profile.web_profile
}
