resource "dagsterplus_service_user" "ci_bot" {
  name        = "ci-bot"
  description = "Service user for CI/CD pipelines"
}
