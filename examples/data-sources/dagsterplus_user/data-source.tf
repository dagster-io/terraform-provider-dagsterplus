data "dagsterplus_user" "alice" {
  email = "alice@example.com"
}

output "alice_role" {
  value = data.dagsterplus_user.alice.role
}
