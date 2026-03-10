resource "dagsterplus_user_token" "ci" {
  name = "ci-pipeline-token"
}

output "user_token_value" {
  value     = dagsterplus_user_token.ci.token
  sensitive = true
}
