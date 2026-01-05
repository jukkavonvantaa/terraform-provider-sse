data "sse_tenant_controls_profiles" "all" {}

output "all_tenant_controls_profiles" {
  value = data.sse_tenant_controls_profiles.all.profiles
}
