data "dagsterplus_agent_token" "hybrid_agent" {
  name = "hybrid-agent-token"
}

output "agent_token_id" {
  value = data.dagsterplus_agent_token.hybrid_agent.id
}
