data "dagsterplus_custom_metric" "my_metric" {
  metadata_key = "asset_runtime_seconds"
}

output "custom_metric_id" {
  value = data.dagsterplus_custom_metric.my_metric.id
}
