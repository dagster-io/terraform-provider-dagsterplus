data "dagsterplus_deployments" "all" {}

output "deployment_names" {
  value = [for d in data.dagsterplus_deployments.all.deployments : d.name]
}
