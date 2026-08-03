# Uses the organization's default agent type.
resource "dagsterplus_deployment" "prod" {
  name = "prod"
}

# Pin the agent type so the deployment is served by your own hybrid agents,
# regardless of the organization default. Changing agent_type switches the
# existing deployment in place.
resource "dagsterplus_deployment" "staging" {
  name       = "staging"
  agent_type = "HYBRID"
}

output "prod_deployment_id" {
  value = dagsterplus_deployment.prod.deployment_id
}
