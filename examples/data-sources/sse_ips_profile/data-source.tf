data "sse_ips_profile" "example" {
  name = "Standard IPS Profile"
}

output "ips_profile_id" {
  value = data.sse_ips_profile.example.id
}
