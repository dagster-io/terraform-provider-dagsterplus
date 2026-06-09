# Grant an agent token the AGENT permission on individual deployments.
# Useful when deployments are managed in a separate module from the token
# (e.g. iterating over deployments with for_each).
resource "dagsterplus_agent_token" "ecs_agent" {
  name         = "ecs-agent-token"
  organization = false
}

resource "dagsterplus_agent_token_deployment_grant" "prod" {
  agent_token_id = dagsterplus_agent_token.ecs_agent.id
  deployment     = "prod"
}

resource "dagsterplus_agent_token_deployment_grant" "staging" {
  agent_token_id = dagsterplus_agent_token.ecs_agent.id
  deployment     = "staging"
}
