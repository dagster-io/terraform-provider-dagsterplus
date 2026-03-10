data "dagsterplus_deployment" "prod" {
  name = "prod"
}

output "prod_deployment_type" {
  value = data.dagsterplus_deployment.prod.type
}
