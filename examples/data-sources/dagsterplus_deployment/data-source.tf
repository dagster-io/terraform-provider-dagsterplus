data "dagsterplus_deployment" "prod" {
  name = "prod"
}

output "prod_deployment_id" {
  value = data.dagsterplus_deployment.prod.deployment_id
}
