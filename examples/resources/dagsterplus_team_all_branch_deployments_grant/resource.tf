resource "dagsterplus_team_all_branch_deployments_grant" "data_eng" {
  team_id = dagsterplus_team.data_engineering.id
  grant   = "EDITOR"
}
