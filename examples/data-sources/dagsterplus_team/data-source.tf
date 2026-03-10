data "dagsterplus_team" "data_engineering" {
  name = "data-engineering"
}

output "team_id" {
  value = data.dagsterplus_team.data_engineering.id
}
