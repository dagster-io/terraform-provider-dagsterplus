resource "dagsterplus_service_user_branch_deployments_grant" "ci_bot_prod_branches" {
  service_user_id   = dagsterplus_service_user.ci_bot.id
  parent_deployment = "prod"
  grant             = "EDITOR"
}
