resource "dagsterplus_service_user_all_branch_deployments_grant" "ci_bot" {
  service_user_id = dagsterplus_service_user.ci_bot.id
  grant           = "LAUNCHER"
}
