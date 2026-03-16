data "dagsterplus_users" "all" {}

output "user_emails" {
  value = [for u in data.dagsterplus_users.all.users : u.email]
}
