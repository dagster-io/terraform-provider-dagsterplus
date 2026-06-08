resource "dagsterplus_user_organization_grant" "alice" {
  user_id = dagsterplus_user.alice.id
  grant   = "ADMIN"
}
