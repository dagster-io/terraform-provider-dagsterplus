# Secret available to all full deployments
resource "dagsterplus_secret" "api_key" {
  secret_name          = "MY_API_KEY"
  secret_value         = "super-secret-api-key-value"
  full_deployment_scope = true
}

# Secret scoped to specific code locations
resource "dagsterplus_secret" "db_password" {
  secret_name    = "DB_PASSWORD"
  secret_value   = "my-database-password"
  location_names = ["data-pipeline", "analytics"]
}
