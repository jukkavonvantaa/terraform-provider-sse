data "sse_network_tunnel_groups" "all" {}

output "all_tunnel_groups" {
  value = data.sse_network_tunnel_groups.all.network_tunnel_groups
}
