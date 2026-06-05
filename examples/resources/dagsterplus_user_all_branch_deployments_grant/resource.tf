resource "dagsterplus_user_all_branch_deployments_grant" "alice" {
  user_id = dagsterplus_user.alice.id
  grant   = "LAUNCHER"
}
