resource "dagsterplus_user_branch_deployments_grant" "alice_prod_branches" {
  user_id           = dagsterplus_user.alice.id
  parent_deployment = "prod"
  grant             = "EDITOR"
}
