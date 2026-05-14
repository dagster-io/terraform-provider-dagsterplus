resource "dagsterplus_service_user_organization_grant" "ci_bot" {
  service_user_id = dagsterplus_service_user.ci_bot.id
  grant           = "ADMIN"
}
