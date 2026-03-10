resource "dagsterplus_custom_metric" "cost" {
  metadata_key = "cloud_cost_usd"
  display_name = "Cloud Cost (USD)"
  description  = "Monthly cloud compute cost in USD"
}

resource "dagsterplus_custom_metric" "rows_processed" {
  metadata_key = "rows_processed"
  display_name = "Rows Processed"
}
