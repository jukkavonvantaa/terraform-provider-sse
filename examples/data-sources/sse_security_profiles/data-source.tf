# Fetch all security profiles
data "sse_security_profiles" "all" {}

output "all_profiles" {
  value = data.sse_security_profiles.all.security_profiles
}
