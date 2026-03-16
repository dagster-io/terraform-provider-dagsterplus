data "dagsterplus_version" "prod" {
  deployment = "prod"
}

output "dagster_version" {
  value = data.dagsterplus_version.prod.version
}
