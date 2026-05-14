resource "dagsterplus_team_organization_grant" "data_eng" {
  team_id = dagsterplus_team.data_engineering.id
  grant   = "ADMIN"
}
