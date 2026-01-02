data "sse_applications" "all" {}

output "all_applications" {
  value = data.sse_applications.all.applications
}

