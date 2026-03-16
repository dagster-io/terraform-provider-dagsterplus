data "dagsterplus_teams" "all" {}

output "team_names" {
  value = [for t in data.dagsterplus_teams.all.teams : t.name]
}
