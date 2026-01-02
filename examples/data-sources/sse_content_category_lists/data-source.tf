data "sse_content_category_lists" "all" {}

output "all_content_category_lists" {
  value = data.sse_content_category_lists.all.content_category_lists
}
