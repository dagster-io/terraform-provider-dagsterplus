data "dagsterplus_secret" "my_api_key" {
  secret_name = "MY_API_KEY"
}

output "secret_id" {
  value = data.dagsterplus_secret.my_api_key.id
}
