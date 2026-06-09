resource "dagsterplus_agent_token" "hybrid_agent" {
  name = "hybrid-agent-token"
}

output "agent_token_value" {
  value     = dagsterplus_agent_token.hybrid_agent.token
  sensitive = true
}

# An ECS agent scoped to specific deployments. Agent tokens always carry the
# AGENT permission, so the grant attributes only select which scopes apply.
resource "dagsterplus_agent_token" "ecs_agent" {
  name = "ecs-agent-token"

  # Remove the organization-wide grant that Dagster+ adds by default.
  organization = false

  # Grant the agent on specific full deployments and all of their branch deployments.
  deployment_grants      = ["prod", "staging"]
  all_branch_deployments = true
}
