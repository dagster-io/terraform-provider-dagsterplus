terraform {
  required_providers {
    dagsterplus = {
      source = "dagster-io/dagsterplus"
    }
  }

  backend "local" {
    path = "terraform.tfstate"
  }
}

# Credentials are read from environment variables:
#   DAGSTER_CLOUD_ORGANIZATION  – your org subdomain
#   DAGSTER_CLOUD_API_TOKEN     – your API token
provider "dagsterplus" {}

# Uncomment to test deployment management:
# resource "dagsterplus_deployment" "test" {
#   name = "tf-test"
#   type = "SERVERLESS"
# }

# resource "dagsterplus_user" "colton" {
#   email = "colton@dagsterlabs.com"
#   role  = "EDITOR"
# }

resource "dagsterplus_alert_policy" "asset_health" {
  deployment  = "prod"
  name        = "asset-health-alerts"
  policy_type = "asset"
  enabled     = false

  asset {
    all_assets    = true
    health_status = "degraded"
  }

  notification_service {
    type            = "email"
    email_addresses = ["dennis@dagsterlabs.com"]
  }
}
