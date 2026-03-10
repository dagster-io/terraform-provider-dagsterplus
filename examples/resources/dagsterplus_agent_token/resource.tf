resource "dagsterplus_agent_token" "hybrid_agent" {
  name = "hybrid-agent-token"
}

output "agent_token_value" {
  value     = dagsterplus_agent_token.hybrid_agent.token
  sensitive = true
}
