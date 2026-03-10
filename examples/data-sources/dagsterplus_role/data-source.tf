# Look up a built-in role (VIEWER, LAUNCHER, EDITOR, ADMIN)
data "dagsterplus_role" "viewer" {
  name = "VIEWER"
}

# Look up a custom role created outside Terraform
data "dagsterplus_role" "data_engineer" {
  name = "data-engineer"
}

# Reference the ID in another resource
output "data_engineer_role_id" {
  value = data.dagsterplus_role.data_engineer.id
}
