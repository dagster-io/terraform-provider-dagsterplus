# Asset policy — all assets, health status, email notification
resource "dagsterplus_alert_policy" "asset_health" {
  deployment  = "prod"
  name        = "asset-health-degraded"
  policy_type = "asset"
  enabled     = true

  asset {
    all_assets    = true
    health_status = "degraded"
  }

  notification_service {
    type            = "email"
    email_addresses = ["oncall@example.com"]
  }
}

# Asset policy — specific asset selection, specific events, Slack notification
resource "dagsterplus_alert_policy" "materialization_failures" {
  deployment  = "prod"
  name        = "critical-asset-failures"
  policy_type = "asset"
  enabled     = true

  asset {
    asset_selection = "tag:critical"
    specific_events = ["materialization_failure", "check_error"]
  }

  notification_service {
    type                 = "slack"
    slack_workspace_name = "my-workspace"
    slack_channel_name   = "#data-alerts"
  }
}

# Run policy — all runs, alert on failure, email notification
resource "dagsterplus_alert_policy" "run_failures" {
  deployment  = "prod"
  name        = "run-failures"
  policy_type = "run"
  enabled     = true

  run {
    all_runs   = true
    on_failure = true
  }

  notification_service {
    type            = "email"
    email_addresses = ["oncall@example.com"]
  }
}

# Run policy — specific jobs, timeout alert
resource "dagsterplus_alert_policy" "long_running_jobs" {
  deployment  = "prod"
  name        = "long-running-jobs"
  policy_type = "run"
  enabled     = true

  run {
    job_names        = ["daily_etl", "weekly_report"]
    on_timeout_hours = 4
  }

  notification_service {
    type            = "pagerduty"
    integration_key = "abc123def456"
  }
}

# Code location policy — all locations, Microsoft Teams notification
resource "dagsterplus_alert_policy" "code_location_errors" {
  deployment  = "prod"
  name        = "code-location-errors"
  policy_type = "code_location"
  enabled     = true

  code_location {
    all_locations = true
  }

  notification_service {
    type        = "microsoft_teams"
    webhook_url = "https://my-org.webhook.office.com/webhookb2/..."
  }
}

# Automation policy — all schedules and sensors, min consecutive failures
resource "dagsterplus_alert_policy" "sensor_failures" {
  deployment  = "prod"
  name        = "sensor-failures"
  policy_type = "automation"
  enabled     = true

  automation {
    all_schedules_and_sensors = true
    include_schedules         = false
    include_sensors           = true
    min_consecutive_failures  = 3
  }

  notification_service {
    type            = "email"
    email_addresses = ["oncall@example.com"]
  }
}

# Budget policy — alert when Dagster credit spend exceeds threshold
resource "dagsterplus_alert_policy" "credit_budget" {
  deployment  = "prod"
  name        = "monthly-credit-budget"
  policy_type = "budget"
  enabled     = true

  budget {
    operator    = "greater_than"
    threshold   = 500
    period_days = 30
  }

  notification_service {
    type            = "email"
    email_addresses = ["finance@example.com", "oncall@example.com"]
  }
}

# Code location policy — generic webhook notification (e.g. IncidentIO)
resource "dagsterplus_alert_policy" "code_location_webhook" {
  deployment  = "prod"
  name        = "code-location-errors-webhook"
  policy_type = "code_location"
  enabled     = true

  code_location {
    all_locations = true
  }

  notification_service {
    type = "webhook"
    # The webhook URL's domain must be allowlisted for your organization
    # (a support-gated setting). body_template tokens must be lowercase
    # identifiers, e.g. {{alert_summary}}, or environment variables, e.g.
    # {{env.MY_SECRET}}.
    webhook_url   = "https://api.incident.io/v2/alert_events/http/00000000"
    body_template = "{\"message\": \"{{alert_summary}}\"}"
    headers = {
      Authorization  = "Bearer my-secret-token"
      "Content-Type" = "application/json"
    }
  }
}

# Insight metric policy — deployment-wide metric threshold
resource "dagsterplus_alert_policy" "weekly_credits" {
  deployment  = "prod"
  name        = "weekly-credit-alert"
  policy_type = "insight_metric"
  enabled     = true

  insight_metric {
    metric      = "dagster_credits"
    operator    = "greater_than"
    threshold   = 100
    period_days = 7
  }

  notification_service {
    type            = "slack"
    slack_workspace_name = "my-workspace"
    slack_channel_name   = "#platform-costs"
  }
}
