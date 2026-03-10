# Configure deployment-level settings for the prod deployment
resource "dagsterplus_deployment_settings" "prod" {
  deployment    = "prod"
  settings_json = jsonencode({
    run_queue = {
      max_concurrent_runs = 10
    }
  })
}
