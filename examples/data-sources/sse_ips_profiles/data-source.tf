data "sse_ips_profiles" "all" {}

output "all_ips_profiles" {
  value = data.sse_ips_profiles.all.ips_profiles
}
