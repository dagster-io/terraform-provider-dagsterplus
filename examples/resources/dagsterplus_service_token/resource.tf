resource "dagsterplus_service_user" "ci_bot" {
  name        = "ci-bot"
  description = "Service user for CI/CD pipelines"
}

resource "dagsterplus_service_token" "ci_token" {
  service_user_id = dagsterplus_service_user.ci_bot.id
  description     = "CI pipeline token"
}

# The token value is only available at creation time
output "ci_token_value" {
  value     = dagsterplus_service_token.ci_token.token
  sensitive = true
}
