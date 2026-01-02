data "sse_application_categories" "all" {}

output "all_application_categories" {
  value = data.sse_application_categories.all.application_categories
}
