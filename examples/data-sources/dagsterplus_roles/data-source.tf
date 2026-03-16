data "dagsterplus_roles" "all" {}

output "role_names" {
  value = [for r in data.dagsterplus_roles.all.roles : r.name]
}
