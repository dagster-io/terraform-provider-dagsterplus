resource "dagsterplus_service_user_deployment_grant" "ci_bot_prod" {
  service_user_id = dagsterplus_service_user.ci_bot.id
  deployment      = "prod"
  grant           = "LAUNCHER"
}
