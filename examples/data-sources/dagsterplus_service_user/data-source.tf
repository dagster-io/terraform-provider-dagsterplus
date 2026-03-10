data "dagsterplus_service_user" "ci_bot" {
  name = "ci-bot"
}

output "service_user_id" {
  value = data.dagsterplus_service_user.ci_bot.id
}
