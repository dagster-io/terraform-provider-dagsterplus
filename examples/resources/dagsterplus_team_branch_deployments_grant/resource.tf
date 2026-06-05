resource "dagsterplus_team_branch_deployments_grant" "data_eng_prod_branches" {
  team_id           = dagsterplus_team.data_engineering.id
  parent_deployment = "prod"
  grant             = "EDITOR"
}
