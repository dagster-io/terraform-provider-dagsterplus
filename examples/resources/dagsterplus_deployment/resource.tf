resource "dagsterplus_deployment" "prod" {
  name = "prod"
}

output "prod_deployment_id" {
  value = dagsterplus_deployment.prod.deployment_id
}
