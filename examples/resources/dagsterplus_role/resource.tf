# Deployment-scoped role for data engineers
resource "dagsterplus_role" "data_engineer" {
  name        = "data-engineer"
  role_type   = "deployment"
  description = "Standard data engineer role"
  icon        = "🛠️"

  permissions = [
    "report_asset_events",
    "wipe_assets",
    "toggle_schedules",
    "toggle_sensors",
    "start_and_stop_runs",
    "read_secret_values",
  ]
}

# Organization-scoped read-only role
resource "dagsterplus_role" "read_only" {
  name        = "read-only"
  role_type   = "organization"
  description = "View-only access across the organization"

  permissions = ["read_secret_values"]
}

# Deployment-scoped ops role with full operational access
resource "dagsterplus_role" "ops" {
  name      = "ops"
  role_type = "deployment"

  permissions = [
    "toggle_schedules",
    "toggle_sensors",
    "edit_sensor_cursors",
    "start_and_stop_runs",
    "delete_runs",
    "edit_concurrency_limits",
    "edit_alerts",
    "read_secret_values",
  ]
}
