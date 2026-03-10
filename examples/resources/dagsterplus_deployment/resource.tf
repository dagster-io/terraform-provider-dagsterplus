resource "dagsterplus_deployment" "prod" {
  name = "prod"
  type = "SERVERLESS"
}

resource "dagsterplus_deployment" "hybrid" {
  name = "hybrid-prod"
  type = "HYBRID"
}

output "prod_deployment_id" {
  value = dagsterplus_deployment.prod.id
}
