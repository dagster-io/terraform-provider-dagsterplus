resource "dagsterplus_deployment" "prod" {
  name = "prod"
  type = "PROD"
}

resource "dagsterplus_deployment" "staging" {
  name = "staging"
  type = "BRANCH"
}

output "prod_deployment_id" {
  value = dagsterplus_deployment.prod.id
}
