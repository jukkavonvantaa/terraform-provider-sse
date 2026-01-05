data "sse_tenant_controls_profile" "example" {
  name = "Global Tenant Controls"
}

output "tenant_controls_profile_id" {
  value = data.sse_tenant_controls_profile.example.id
}
