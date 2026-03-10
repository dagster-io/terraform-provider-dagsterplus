# Snowflake external asset connection with a schedule
resource "dagsterplus_external_asset_connection" "snowflake_daily" {
  name               = "snowflake-daily-sync"
  description        = "Daily sync from Snowflake"
  connection_type    = "SNOWFLAKE"
  source_config_yaml = <<-YAML
    account: my_account
    user: my_user
    warehouse: my_warehouse
    database: my_database
    schema: public
  YAML
  cron_string        = "0 6 * * *"
  timezone           = "America/New_York"
}

# BigQuery connection without a schedule
resource "dagsterplus_external_asset_connection" "bigquery_sync" {
  name               = "bigquery-sync"
  connection_type    = "BIGQUERY"
  source_config_yaml = <<-YAML
    project: my-gcp-project
    dataset: my_dataset
  YAML
}
