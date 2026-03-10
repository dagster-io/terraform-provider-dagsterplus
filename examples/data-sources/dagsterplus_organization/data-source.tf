data "dagsterplus_organization" "current" {}

output "org_name" {
  value = data.dagsterplus_organization.current.name
}

output "org_status" {
  value = data.dagsterplus_organization.current.status
}

output "github_account" {
  value = data.dagsterplus_organization.current.github_account_name
}
