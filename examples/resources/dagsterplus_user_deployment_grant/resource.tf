resource "dagsterplus_user_deployment_grant" "alice_prod" {
  user_id    = dagsterplus_user.alice.id
  deployment = "prod"
  grant      = "EDITOR"
}
