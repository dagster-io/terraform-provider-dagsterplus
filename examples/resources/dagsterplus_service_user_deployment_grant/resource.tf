resource "dagsterplus_service_user" "ci_bot" {
  name        = "ci-bot"
  description = "Service user for CI/CD pipelines"
}

resource "dagsterplus_service_user_deployment_grant" "ci_bot_prod" {
  service_user_id = dagsterplus_service_user.ci_bot.id
  deployment      = "prod"
  grant           = "LAUNCHER"
}

# Grant with per-code-location overrides
resource "dagsterplus_service_user_deployment_grant" "ci_bot_staging" {
  service_user_id = dagsterplus_service_user.ci_bot.id
  deployment      = "staging"
  grant           = "VIEWER"

  location_grants {
    location_name = "data-platform"
    grant         = "EDITOR"
  }
}
